'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { motion, AnimatePresence } from 'framer-motion';
import {
    ArrowLeft,
    Clock,
    Plus,
    Save,
    Trash2,
    MessageCircle,
    Mail,
    Phone,
    Bell,
    CheckCircle2,
    Settings,
    MoreVertical,
    Play,
    Pause,
    Moon
} from 'lucide-react';

// Backend Models (inferred)
interface Campaign {
    id: string;
    name: string;
    trigger_type: string;
    is_active: boolean;
    quiet_hours_enabled: boolean;
    steps: CampaignStep[];
}

interface CampaignStep {
    id?: string;
    step_order: number;
    delay_hours: number;
    channel: string;
    template_id?: string;
    config?: any;
}

const DEFAULT_STEPS = [
    {
        id: 'trigger',
        type: 'trigger',
        label: 'Nouvelle transaction sans justificatif',
        delay: 0,
        icon: Bell,
        color: 'bg-[#1A1A1A] text-white'
    },
    {
        id: 'step-1',
        type: 'action',
        channel: 'combined', // mapped to email + whatsapp in backend logic or multi-step
        label: 'Impact Immédiat',
        description: 'Email formel + WhatsApp de proximité',
        delay: 0,
        icon: MessageCircle,
        color: 'bg-[#1A4D2E] text-white',
        content: {
            whatsapp: "Bonjour [Client], document manquant pour [Montant]. Une photo ?",
            email: "Relevé des pièces manquantes - Fiducia"
        }
    },
    {
        id: 'step-2',
        type: 'action',
        channel: 'voice',
        label: 'Relance Vocale IA',
        description: 'Message vocal cloné (Collab. assigné)',
        delay: 48, // 48h = 2 days
        icon: Phone,
        color: 'bg-[#4F2830] text-white',
        content: {
            script: "Bonjour, c'est [Nom] du cabinet..."
        }
    },
    {
        id: 'step-3',
        type: 'action',
        channel: 'notification',
        label: 'Escalade Interne',
        description: 'Notification au collaborateur',
        delay: 120, // 5 days total
        icon: Bell,
        color: 'bg-amber-600 text-white',
        content: {
            msg: "Intervention manuelle requise"
        }
    }
];

export default function CampaignsPage() {
    const [campaignId, setCampaignId] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [isSaving, setIsSaving] = useState(false);
    const [quietHours, setQuietHours] = useState(true);
    const [isActive, setIsActive] = useState(true);
    const [steps, setSteps] = useState<any[]>(DEFAULT_STEPS);

    // Load Campaign on Mount
    useEffect(() => {
        fetchCampaigns();
    }, []);

    const fetchCampaigns = async () => {
        try {
            const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/campaigns`);
            if (!res.ok) throw new Error('Failed to fetch campaigns');

            const campaigns: Campaign[] = await res.json();
            if (campaigns && campaigns.length > 0) {
                // Determine the most relevant campaign (e.g., the first active one or the latest)
                const campaign = campaigns[0];
                setCampaignId(campaign.id);
                setQuietHours(campaign.quiet_hours_enabled);
                setIsActive(campaign.is_active);

                // Map backend steps to frontend steps
                if (campaign.steps && campaign.steps.length > 0) {
                    const mappedSteps = [
                        DEFAULT_STEPS[0], // Always keep the trigger
                        ...campaign.steps.map((s, i) => mapBackendStepToFrontend(s, i))
                    ];
                    setSteps(mappedSteps);
                }
            }
        } catch (err) {
            console.error('Error loading campaigns:', err);
        } finally {
            setIsLoading(false);
        }
    };

    const mapBackendStepToFrontend = (s: CampaignStep, index: number) => {
        // Simple mapping logic - in a real app this would be more robust
        let icon = MessageCircle;
        let color = 'bg-[#1A4D2E] text-white';
        let label = 'Action';
        let description = '';

        if (s.channel === 'voice') {
            icon = Phone;
            color = 'bg-[#4F2830] text-white';
            label = 'Relance Vocale IA';
            description = 'Message vocal cloné';
        } else if (s.channel === 'notification') {
            icon = Bell;
            color = 'bg-amber-600 text-white';
            label = 'Escalade Interne';
            description = 'Notification au collaborateur';
        } else {
            // Default / Combined
            label = 'Impact Immédiat';
            description = 'Email + WhatsApp';
        }

        return {
            id: s.id || `step-${index}`,
            type: 'action',
            channel: s.channel,
            label,
            description,
            delay: s.delay_hours,
            icon,
            color,
            content: s.config
        };
    };

    const saveCampaign = async () => {
        setIsSaving(true);
        try {
            // Filter out the trigger step (index 0)
            const backendSteps = steps.filter(s => s.type === 'action').map((s, i) => ({
                step_order: i + 1,
                delay_hours: s.delay,
                channel: s.channel === 'combined' ? 'whatsapp' : s.channel, // Simplified: combined -> whatsapp for now
                template_id: 'default',
                config: s.content
            }));

            const payload = {
                name: "Smart Sequence v2",
                trigger_type: "on_pending",
                is_active: isActive,
                quiet_hours_enabled: quietHours,
                steps: backendSteps
            };

            const url = campaignId
                ? `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/campaigns/${campaignId}`
                : `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/campaigns`;

            const method = campaignId ? 'PATCH' : 'POST';

            const res = await fetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (!res.ok) throw new Error('Failed to save');

            const saved = await res.json();
            if (!campaignId) setCampaignId(saved.id);

            // Simple "toast" log
            console.log('Campaign saved successfully');

        } catch (err) {
            console.error('Failed to save campaign:', err);
        } finally {
            setIsSaving(false);
        }
    };

    const formatDelay = (hours: number) => {
        if (hours === 0) return 'Immédiat';
        if (hours >= 24) return `J+${hours / 24}`;
        return `${hours}h`;
    };

    return (
        <div className="min-h-screen bg-[#F9F8F6] text-[#1A1A1A] font-sans selection:bg-[#4F2830]/20">
            {/* Navbar */}
            <nav className="sticky top-0 z-50 bg-[#F9F8F6]/80 backdrop-blur-md border-b border-[#1A1A1A]/5 px-4 md:px-8 h-16 md:h-20 flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <Link href="/dashboard" className="p-2 -ml-2 text-[#1A1A1A]/40 hover:text-[#1A1A1A] transition-colors">
                        <ArrowLeft size={20} />
                    </Link>
                    <span className="text-xl font-serif font-bold tracking-tight">Moteur de Relance</span>
                </div>
                <div className="flex items-center gap-4">
                    <div className={`hidden md:flex items-center gap-2 px-3 py-1.5 border border-[#1A1A1A]/5 rounded-full text-xs font-medium transition-colors ${isActive ? 'bg-white text-[#1A1A1A]/60' : 'bg-red-50 text-red-600 border-red-100'}`}>
                        <div className={`w-2 h-2 rounded-full ${isActive ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`} />
                        {isActive ? 'Système Actif' : 'Système Inactif'}
                    </div>
                </div>
            </nav>

            <main className="max-w-7xl mx-auto px-4 md:px-6 py-8 md:py-12">
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                    {/* Left Panel: Settings & Stats */}
                    <div className="space-y-6">
                        <div className="bg-white p-6 rounded-2xl border border-[#1A1A1A]/5 shadow-sm">
                            <h2 className="font-serif text-xl mb-4">Campagne Active</h2>
                            <div className="flex items-center justify-between p-4 bg-[#F9F8F6] rounded-xl border border-[#1A1A1A]/5 mb-6">
                                <div>
                                    <div className="font-bold text-sm">Smart Sequence v2</div>
                                    <div className="text-xs text-[#1A1A1A]/60">{steps.length - 1} étapes</div>
                                </div>
                                <div
                                    onClick={() => setIsActive(!isActive)}
                                    className={`w-10 h-6 rounded-full relative cursor-pointer transition-colors ${isActive ? 'bg-[#1A4D2E]' : 'bg-[#1A1A1A]/20'}`}
                                >
                                    <div className={`absolute top-1 w-4 h-4 bg-white rounded-full shadow-sm transition-all ${isActive ? 'right-1' : 'left-1'}`} />
                                </div>
                            </div>

                            <h3 className="text-sm font-bold uppercase tracking-wider text-[#1A1A1A]/40 mb-3">Paramètres "African-First"</h3>

                            <div className="flex items-center justify-between py-3 border-b border-[#1A1A1A]/5">
                                <div className="flex items-center gap-3">
                                    <div className="w-8 h-8 rounded-full bg-[#1A1A1A]/5 flex items-center justify-center text-[#1A1A1A]">
                                        <Moon size={14} />
                                    </div>
                                    <div>
                                        <div className="text-sm font-medium">Quiet Hours</div>
                                        <div className="text-xs text-[#1A1A1A]/40">Pas d'envoi après 18h & WE</div>
                                    </div>
                                </div>
                                <button
                                    onClick={() => setQuietHours(!quietHours)}
                                    className={`w-10 h-6 rounded-full relative transition-colors ${quietHours ? 'bg-[#1A4D2E]' : 'bg-[#1A1A1A]/20'}`}
                                >
                                    <div className={`absolute top-1 w-4 h-4 bg-white rounded-full shadow-sm transition-all ${quietHours ? 'right-1' : 'left-1'}`} />
                                </button>
                            </div>

                            <div className="flex items-center justify-between py-3">
                                <div className="flex items-center gap-3">
                                    <div className="w-8 h-8 rounded-full bg-[#1A1A1A]/5 flex items-center justify-center text-[#1A1A1A]">
                                        <MessageCircle size={14} />
                                    </div>
                                    <div>
                                        <div className="text-sm font-medium">IA Variation</div>
                                        <div className="text-xs text-[#1A1A1A]/40">Anti-spam WhatsApp</div>
                                    </div>
                                </div>
                                <div className="text-xs font-bold text-[#1A4D2E]">ACTIF</div>
                            </div>
                        </div>

                        {/* Performance Stats */}
                        <div className="bg-[#1A1A1A] text-white p-6 rounded-2xl shadow-xl relative overflow-hidden">
                            <div className="absolute top-0 right-0 p-32 bg-[#4F2830] rounded-full mix-blend-screen blur-3xl opacity-20 -mr-10 -mt-10" />

                            <h2 className="font-serif text-xl mb-6 relative z-10">Performance</h2>
                            <div className="grid grid-cols-2 gap-4 relative z-10">
                                <div>
                                    <div className="text-white/40 text-xs mb-1">Taux d'ouverture</div>
                                    <div className="text-2xl font-serif">94%</div>
                                    <div className="text-green-400 text-xs flex items-center gap-1">
                                        <ArrowLeft size={10} className="rotate-45" /> +2.4%
                                    </div>
                                </div>
                                <div>
                                    <div className="text-white/40 text-xs mb-1">Récupération</div>
                                    <div className="text-2xl font-serif">78%</div>
                                    <div className="text-green-400 text-xs flex items-center gap-1">
                                        <ArrowLeft size={10} className="rotate-45" /> +5.1%
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Right Panel: Timeline Builder */}
                    <div className="lg:col-span-2">
                        <div className="bg-white rounded-2xl border border-[#1A1A1A]/5 shadow-sm overflow-hidden flex flex-col min-h-[600px]">
                            <div className="p-6 border-b border-[#1A1A1A]/5 flex justify-between items-center">
                                <h1 className="font-serif text-2xl">Séquence de Relance</h1>
                                <button
                                    onClick={saveCampaign}
                                    disabled={isSaving}
                                    className="px-4 py-2 bg-[#1A1A1A] text-white rounded-lg text-sm font-medium hover:bg-[#1A4D2E] transition-colors flex items-center gap-2 disabled:opacity-50"
                                >
                                    <Save size={16} />
                                    {isSaving ? 'Sauvegarde...' : 'Sauvegarder'}
                                </button>
                            </div>

                            <div className="flex-1 p-8 bg-[#F9F8F6]/50 relative">
                                {isLoading ? (
                                    <div className="absolute inset-0 flex items-center justify-center">
                                        <div className="animate-spin w-8 h-8 border-2 border-[#1A1A1A] border-t-transparent rounded-full" />
                                    </div>
                                ) : (
                                    <div className="space-y-8 relative">
                                        {/* Vertical Line */}
                                        <div className="absolute left-8 md:left-12 top-0 bottom-0 w-px bg-[#1A1A1A]/10" />

                                        {steps.map((step, index) => (
                                            <motion.div
                                                key={step.id}
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: index * 0.1 }}
                                                className="relative flex gap-6 md:gap-8 group"
                                            >
                                                {/* Timestamp/Delay Badge */}
                                                <div className="w-16 flex-shrink-0 flex flex-col items-center pt-2 relative z-10">
                                                    <div className={`w-8 h-8 rounded-full flex items-center justify-center shadow-lg transition-transform group-hover:scale-110 ${step.color}`}>
                                                        <step.icon size={14} />
                                                    </div>
                                                    {index < steps.length - 1 && (
                                                        <div className="mt-2 py-1 px-2 bg-white border border-[#1A1A1A]/10 rounded-md text-[10px] font-bold text-[#1A1A1A]/60 shadow-sm z-20">
                                                            {index === 0 ? formatDelay(steps[index + 1].delay) : formatDelay(steps[index + 1].delay - step.delay)}
                                                        </div>
                                                    )}
                                                </div>

                                                {/* Content Card */}
                                                <div className="flex-1 bg-white p-5 rounded-xl border border-[#1A1A1A]/5 shadow-sm hover:shadow-md transition-all group-hover:border-[#1A1A1A]/20">
                                                    <div className="flex justify-between items-start mb-2">
                                                        <div>
                                                            <h3 className="font-bold text-[#1A1A1A]">{step.label}</h3>
                                                            <p className="text-sm text-[#1A1A1A]/60">{step.description}</p>
                                                        </div>
                                                        {step.type !== 'trigger' && (
                                                            <button className="text-[#1A1A1A]/20 hover:text-red-500 transition-colors">
                                                                <Trash2 size={16} />
                                                            </button>
                                                        )}
                                                    </div>

                                                    {/* Stop Condition Badge (Between Steps logic visual) */}
                                                    <div className="mt-4 flex flex-wrap gap-2">
                                                        <div className="inline-flex items-center gap-1.5 px-2 py-1 bg-green-50 text-green-700 rounded-md text-[10px] font-medium border border-green-100">
                                                            <CheckCircle2 size={10} />
                                                            Stop si validé
                                                        </div>
                                                        {step.channel === 'combined' && (
                                                            <span className="text-[10px] text-[#1A1A1A]/40 flex items-center p-1">
                                                                + Variante IA activée
                                                            </span>
                                                        )}
                                                    </div>
                                                </div>
                                            </motion.div>
                                        ))}

                                        {/* Add Step Button */}
                                        <div className="relative flex gap-8">
                                            <div className="w-16 flex justify-center">
                                                <button className="w-8 h-8 rounded-full bg-white border border-[#1A1A1A]/20 flex items-center justify-center text-[#1A1A1A]/40 hover:text-[#1A1A1A] hover:border-[#1A1A1A] transition-all">
                                                    <Plus size={16} />
                                                </button>
                                            </div>
                                            <div className="flex-1 py-1 text-sm text-[#1A1A1A]/40 italic">
                                                Ajouter une étape...
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
}
