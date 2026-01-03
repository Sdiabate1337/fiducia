package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
)

type CampaignEngine struct {
	pool          *pgxpool.Pool
	campaignRepo  *repository.CampaignRepository
	lineRepo      *repository.PendingLineRepository
	executionRepo *repository.CampaignExecutionRepository
	voiceSvc      *VoiceService
}

func NewCampaignEngine(pool *pgxpool.Pool, campaignRepo *repository.CampaignRepository, lineRepo *repository.PendingLineRepository, executionRepo *repository.CampaignExecutionRepository, voiceSvc *VoiceService) *CampaignEngine {
	return &CampaignEngine{
		pool:          pool,
		campaignRepo:  campaignRepo,
		lineRepo:      lineRepo,
		executionRepo: executionRepo,
		voiceSvc:      voiceSvc,
	}
}

// CheckStopCondition verifies if a campaign should stop for a pending line
func (e *CampaignEngine) CheckStopCondition(ctx context.Context, lineID uuid.UUID) (bool, models.StopReason, error) {
	line, err := e.lineRepo.GetByID(ctx, lineID)
	if err != nil {
		return false, "", err
	}

	// 1. Validated (Manual or Auto)
	if line.Status == models.StatusValidated {
		return true, models.StopManualValidated, nil
	}

	// 2. Client Refusal (Rejected)
	if line.Status == models.StatusRejected {
		return true, models.StopClientRefusal, nil
	}

	// 3. Document received but not validated yet (StatusReceived)
	// If the rule is "Stop if doc received", then stop.
	if line.Status == models.StatusReceived {
		return true, models.StopOCRValidated, nil // Using this reason for now
	}

	return false, "", nil
}

// IsQuietHours returns true if it's currently quiet hours (after 18h or weekend)
func (e *CampaignEngine) IsQuietHours(t time.Time) bool {
	// Weekend (Saturday=6, Sunday=0)
	if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		return true
	}

	// After 18h or before 8h
	hour := t.Hour()
	if hour >= 18 || hour < 8 {
		return true
	}

	return false
}

// ExecuteCampaignCycle is the main entry point called by the worker
func (e *CampaignEngine) ExecuteCampaignCycle(ctx context.Context) error {
	slog.Info("Running campaign cycle...")

	// 1. Get all active campaigns
	// For MVP we can just iterate over all cabinets via some method or just assume one cabinet/campaign context if triggered per cabinet.
	// BETTER: Just list ALL active campaigns across all cabinets. But repo.List takes cabinetID.
	// Hack for MVP: Hardcode the Demo Cabinet ID or add a ListAllActive to repo.
	demoCabinetID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	campaigns, err := e.campaignRepo.List(ctx, demoCabinetID)
	if err != nil {
		return err
	}

	for _, campaign := range campaigns {
		if !campaign.IsActive {
			continue
		}

		// 1. Enroll New Lines (Trigger: OnPending)
		if campaign.TriggerType == models.TriggerOnPending {
			unrolledIDs, err := e.executionRepo.FindUnenrolledLines(ctx, campaign.ID, campaign.CabinetID)
			if err != nil {
				slog.Error("failed to find unenrolled lines", "campaign", campaign.Name, "error", err)
				continue
			}
			for _, lineID := range unrolledIDs {
				// Create Execution
				exec := models.CampaignExecution{
					CampaignID:    campaign.ID,
					PendingLineID: lineID,
					Status:        models.ExecStatusPending,
					// Schedule first step immediately or after delay? Step 1 usually has delay 0.
					NextStepScheduledAt: ptrTo(time.Now()),
				}
				if err := e.executionRepo.Create(ctx, &exec); err != nil {
					slog.Error("failed to enroll line", "lineID", lineID, "error", err)
				} else {
					slog.Info("Enrolled line in campaign", "lineID", lineID, "campaign", campaign.Name)
				}
			}
		}
	}

	// 2. Process Active Executions
	executions, err := e.executionRepo.FindActive(ctx)
	if err != nil {
		return err
	}

	for _, ex := range executions {
		// Load Campaign info to check QuietHours
		camp, err := e.campaignRepo.GetByID(ctx, ex.CampaignID)
		if err != nil || camp == nil {
			continue
		}

		// Check Stop Condition
		shouldStop, reason, err := e.CheckStopCondition(ctx, ex.PendingLineID)
		if err != nil {
			slog.Error("failed to check stop condition", "execID", ex.ID, "error", err)
			continue
		}
		if shouldStop {
			ex.Status = models.ExecStatusStopped
			ex.StopReason = &reason
			e.executionRepo.Update(ctx, &ex)
			slog.Info("Campaign stopped", "execID", ex.ID, "reason", reason)
			continue
		}

		// Check Schedule
		if ex.NextStepScheduledAt == nil || time.Now().Before(*ex.NextStepScheduledAt) {
			continue
		}

		// Check Quiet Hours
		if camp.QuietHoursEnabled && e.IsQuietHours(time.Now()) {
			slog.Info("Skipping execution due to Quiet Hours", "execID", ex.ID)
			continue
		}

		// Execute Step
		if err := e.executeNextStep(ctx, &ex, *camp); err != nil {
			slog.Error("failed to execute step", "execID", ex.ID, "error", err)
		}
	}

	return nil
}

func (e *CampaignEngine) executeNextStep(ctx context.Context, ex *models.CampaignExecution, campaign models.Campaign) error {
	// Find the step matching current_step_order + 1 (if starting) or next
	// Logic: ex.CurrentStepOrder is the LAST executed step. So look for CurrentStepOrder + 1
	nextOrder := ex.CurrentStepOrder + 1

	// Find step in campaign.Steps
	var nextStep *models.CampaignStep
	for _, s := range campaign.Steps {
		if s.StepOrder == nextOrder {
			nextStep = &s
			break
		}
	}

	if nextStep == nil {
		// No more steps -> Complete
		ex.Status = models.ExecStatusCompleted
		stopReason := models.StopCompleted
		ex.StopReason = &stopReason
		return e.executionRepo.Update(ctx, ex)
	}

	// Execute Action
	slog.Info("EXECUTE ACTION", "channel", nextStep.Channel, "lineID", ex.PendingLineID)
	// Here we would call EmailService or WhatsAppService
	// e.g. e.msgSvc.Send(..., nextStep.TemplateID, ...)

	// Advance State
	ex.CurrentStepOrder = nextOrder
	ex.LastStepExecutedAt = ptrTo(time.Now())

	// Schedule Next Step (if any)
	futureOrder := nextOrder + 1
	var futureStep *models.CampaignStep
	for _, s := range campaign.Steps {
		if s.StepOrder == futureOrder {
			futureStep = &s
			break
		}
	}

	if futureStep != nil {
		// Schedule for Now + Delay
		delay := time.Duration(futureStep.DelayHours) * time.Hour
		nextTime := time.Now().Add(delay)
		ex.NextStepScheduledAt = &nextTime
		ex.Status = models.ExecStatusRunning
	} else {
		// Finished
		ex.Status = models.ExecStatusCompleted
		ex.NextStepScheduledAt = nil
		stopReason := models.StopCompleted
		ex.StopReason = &stopReason
	}

	return e.executionRepo.Update(ctx, ex)
}

func ptrTo[T any](v T) *T {
	return &v
}
