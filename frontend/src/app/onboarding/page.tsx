'use client';

import { useState, useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    ArrowRight,
    Building2,
    Database,
    Mic,
    CheckCircle2,
    Upload,
    Play,
    StopCircle,
    Loader2,
    Users
} from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { cn } from '@/lib/utils';
import { useAuth } from '@/context/AuthContext';

// --- Sub-Components ---

const StepIndicator = ({ currentStep, totalSteps }: { currentStep: number; totalSteps: number }) => (
    <div className="flex items-center gap-2 mb-8">
        {[...Array(totalSteps)].map((_, i) => (
            <div key={i} className="flex items-center">
                <div
                    className={cn(
                        "w-2.5 h-2.5 rounded-full transition-all duration-500",
                        i + 1 <= currentStep ? "bg-[#1A4D2E] w-8" : "bg-[#1A1A1A]/10"
                    )}
                />
            </div>
        ))}
    </div>
);

const IdentityStep = ({ onNext }: { onNext: () => void }) => {
    const { user, token } = useAuth();
    const [name, setName] = useState('');
    const [logo, setLogo] = useState<File | null>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);

    // Pre-fill name if available in user context (e.g. from register)
    useEffect(() => {
        const fetchCabinet = async () => {
            if (!user?.cabinet_id || !token) return;
            try {
                const res = await fetch(`http://localhost:8080/api/v1/cabinets/${user.cabinet_id}`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                if (res.ok) {
                    const data = await res.json();
                    if (data.name) setName(data.name);
                }
            } catch (err) {
                console.error("Failed to fetch cabinet", err);
            }
        };
        fetchCabinet();
    }, [user, token]);

    const handleSubmit = async () => {
        if (!name) return;
        setIsSubmitting(true);

        try {
            // Update Cabinet Name
            const res = await fetch(`http://localhost:8080/api/v1/cabinets/${user?.cabinet_id}`, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ name: name })
            });

            if (!res.ok) throw new Error('Failed to update cabinet');

            // TODO: Upload Logo if present (separate endpoint or Multipart)

            onNext();
        } catch (err) {
            console.error(err);
            // Allow proceed for MVP even if error
            onNext();
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} className="space-y-6">
            <div className="text-center md:text-left">
                <h2 className="text-2xl font-serif font-bold text-[#1A1A1A]">Identité du Cabinet</h2>
                <p className="text-[#1A1A1A]/60">Personnalisons l'expérience pour vos clients.</p>
            </div>

            <div className="space-y-4">
                <div>
                    <label className="block text-xs font-bold uppercase tracking-widest text-[#1A1A1A]/40 mb-2">Nom du Cabinet</label>
                    <input
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        placeholder="Ex: Cabinet Martin & Associés"
                        className="w-full bg-white border border-[#1A1A1A]/10 rounded-xl px-5 py-4 font-medium text-[#1A1A1A] outline-none focus:border-[#1A4D2E]"
                    />
                </div>

                <div>
                    <label className="block text-xs font-bold uppercase tracking-widest text-[#1A1A1A]/40 mb-2">Logo (Optionnel)</label>
                    <div className="border-2 border-dashed border-[#1A1A1A]/10 rounded-xl p-8 flex flex-col items-center justify-center text-center hover:bg-[#F9F8F6] transition-colors cursor-pointer relative">
                        <input
                            type="file"
                            className="absolute inset-0 opacity-0 cursor-pointer"
                            onChange={(e) => setLogo(e.target.files?.[0] || null)}
                        />
                        {logo ? (
                            <div className="flex flex-col items-center">
                                <div className="w-16 h-16 bg-[#1A4D2E]/10 rounded-full flex items-center justify-center text-[#1A4D2E] mb-2">
                                    <CheckCircle2 size={24} />
                                </div>
                                <span className="text-sm font-medium">{logo.name}</span>
                            </div>
                        ) : (
                            <>
                                <Upload size={24} className="text-[#1A1A1A]/20 mb-2" />
                                <span className="text-sm text-[#1A1A1A]/40 font-medium">Glisser ou cliquer pour uploader</span>
                            </>
                        )}
                    </div>
                </div>
            </div>

            <button
                disabled={!name || isSubmitting}
                onClick={handleSubmit}
                className="w-full py-4 rounded-xl bg-[#1A1A1A] text-white font-bold tracking-wide uppercase shadow-lg flex items-center justify-center gap-2 hover:bg-[#1A4D2E] transition-colors disabled:opacity-50 disabled:pointer-events-none"
            >
                {isSubmitting ? (
                    'Enregistrement...'
                ) : (
                    <>Suivant <ArrowRight size={18} /></>
                )}
            </button>
        </motion.div>
    );
};

const ClientsStep = ({ onNext }: { onNext: () => void }) => {
    const { user, token } = useAuth();
    const [uploading, setUploading] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

    const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        setUploading(true);
        const formData = new FormData();
        formData.append('file', file);

        try {
            const res = await fetch(`http://localhost:8080/api/v1/cabinets/${user?.cabinet_id}/import/clients`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                body: formData
            });

            if (!res.ok) throw new Error('Upload failed');
            onNext();
        } catch (err) {
            console.error("Upload failed", err);
            setUploading(false);
            alert("Erreur lors de l'import. Vérifiez le format du fichier ou réessayez.");
        }
    };

    return (
        <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} className="space-y-6">
            <div className="text-center md:text-left">
                <h2 className="text-2xl font-serif font-bold text-[#1A1A1A]">Carnet d'Adresses</h2>
                <p className="text-[#1A1A1A]/60">Importez vos clients (Nom, Email, Tel) pour activer le matching automatique.</p>
            </div>

            <div
                onClick={() => fileInputRef.current?.click()}
                className="border-2 border-dashed border-[#1A1A1A]/10 rounded-2xl p-12 flex flex-col items-center justify-center text-center hover:border-[#1A4D2E]/30 hover:bg-[#1A4D2E]/5 transition-all cursor-pointer group"
            >
                <input
                    type="file"
                    ref={fileInputRef}
                    onChange={handleFileSelect}
                    className="hidden"
                    accept=".csv,.txt,.xlsx"
                />

                {uploading ? (
                    <div className="space-y-4">
                        <Loader2 size={48} className="text-[#1A4D2E] animate-spin mx-auto" />
                        <div>
                            <p className="font-bold text-[#1A1A1A]">Importation...</p>
                            <p className="text-sm text-[#1A1A1A]/40">Création des fiches clients</p>
                        </div>
                    </div>
                ) : (
                    <>
                        <div className="w-20 h-20 bg-white rounded-full shadow-lg flex items-center justify-center text-[#1A1A1A]/20 group-hover:text-[#1A4D2E] group-hover:scale-110 transition-all mb-6">
                            <Users size={32} />
                        </div>
                        <h3 className="font-bold text-lg mb-2">Fichier Clients / Facturation</h3>
                        <p className="text-[#1A1A1A]/40 text-sm max-w-xs mx-auto">
                            Format: Nom, Email, Téléphone (Export CRM ou Facturation)
                        </p>
                        <button className="mt-8 px-6 py-2 bg-white border border-[#1A1A1A]/10 rounded-full text-sm font-bold shadow-sm group-hover:shadow-md transition-all">
                            Sélectionner le fichier
                        </button>
                    </>
                )}
            </div>

            <button
                onClick={onNext}
                className="w-full py-3 rounded-xl border border-[#1A1A1A]/10 text-[#1A1A1A]/40 font-bold tracking-wide uppercase text-xs hover:text-[#1A1A1A] transition-colors"
            >
                Je n'ai pas de fichier clients (Passer)
            </button>
        </motion.div>
    );
};

const DataStep = ({ onNext }: { onNext: () => void }) => {
    const { user, token } = useAuth();
    const [uploading, setUploading] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

    const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        setUploading(true);
        const formData = new FormData();
        formData.append('file', file);

        try {
            const res = await fetch(`http://localhost:8080/api/v1/cabinets/${user?.cabinet_id}/import/csv`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                body: formData
            });

            if (!res.ok) throw new Error('Upload failed');
            onNext();
        } catch (err) {
            console.error("Upload failed", err);
            setUploading(false);
            alert("Erreur lors de l'import. Vérifiez le format du fichier ou réessayez.");
        }
    };

    return (
        <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} className="space-y-6">
            <div className="text-center md:text-left">
                <h2 className="text-2xl font-serif font-bold text-[#1A1A1A]">Le Pouls Financier</h2>
                <p className="text-[#1A1A1A]/60">Importez votre Grand Livre (FEC ou Sage) pour activer le dashboard.</p>
            </div>

            <div
                onClick={() => fileInputRef.current?.click()}
                className="border-2 border-dashed border-[#1A1A1A]/10 rounded-2xl p-12 flex flex-col items-center justify-center text-center hover:border-[#1A4D2E]/30 hover:bg-[#1A4D2E]/5 transition-all cursor-pointer group"
            >
                <input
                    type="file"
                    ref={fileInputRef}
                    onChange={handleFileSelect}
                    className="hidden"
                    accept=".csv,.txt,.xlsx"
                />

                {uploading ? (
                    <div className="space-y-4">
                        <Loader2 size={48} className="text-[#1A4D2E] animate-spin mx-auto" />
                        <div>
                            <p className="font-bold text-[#1A1A1A]">Analyse en cours...</p>
                            <p className="text-sm text-[#1A1A1A]/40">Détection des factures impayées</p>
                        </div>
                    </div>
                ) : (
                    <>
                        <div className="w-20 h-20 bg-white rounded-full shadow-lg flex items-center justify-center text-[#1A1A1A]/20 group-hover:text-[#1A4D2E] group-hover:scale-110 transition-all mb-6">
                            <Database size={32} />
                        </div>
                        <h3 className="font-bold text-lg mb-2">Glisser votre fichier FEC ici</h3>
                        <p className="text-[#1A1A1A]/40 text-sm max-w-xs mx-auto">
                            Formats acceptés: .csv, .txt, .xlsx. Traitement instantané et sécurisé.
                        </p>
                        <button className="mt-8 px-6 py-2 bg-white border border-[#1A1A1A]/10 rounded-full text-sm font-bold shadow-sm group-hover:shadow-md transition-all">
                            Ou parcourir les fichiers
                        </button>
                    </>
                )}
            </div>

            <button
                onClick={onNext}
                className="w-full py-3 rounded-xl border border-[#1A1A1A]/10 text-[#1A1A1A]/40 font-bold tracking-wide uppercase text-xs hover:text-[#1A1A1A] transition-colors"
            >
                Passer cette étape
            </button>
        </motion.div>
    );
};

const VoiceStep = ({ onNext }: { onNext: () => void }) => {
    const [isRecording, setIsRecording] = useState(false);
    const [isCalibrating, setIsCalibrating] = useState(false);

    const handleRecord = () => {
        if (isRecording) {
            setIsRecording(false);
            setIsCalibrating(true);
            setTimeout(() => {
                setIsCalibrating(false);
                onNext();
            }, 3000);
        } else {
            setIsRecording(true);
        }
    };

    return (
        <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} className="space-y-6">
            <div className="text-center md:text-left">
                <h2 className="text-2xl font-serif font-bold text-[#1A1A1A]">Votre Double Numérique</h2>
                <p className="text-[#1A1A1A]/60">Calibrez l'IA pour qu'elle relance vos clients avec votre voix.</p>
            </div>

            <div className="bg-white rounded-2xl border border-[#1A1A1A]/5 p-8 text-center relative overflow-hidden">
                {isCalibrating ? (
                    <div className="py-12 space-y-4">
                        <div className="w-24 h-24 mx-auto bg-[#1A4D2E]/5 rounded-full flex items-center justify-center relative">
                            <div className="absolute inset-0 border-4 border-[#1A4D2E]/20 rounded-full animate-ping" />
                            <div className="w-16 h-16 bg-[#1A4D2E] rounded-full flex items-center justify-center text-white">
                                <Mic size={32} />
                            </div>
                        </div>
                        <h3 className="font-serif text-xl font-bold text-[#1A1A1A]">Calibrage...</h3>
                        <p className="text-sm text-[#1A1A1A]/60">Création de votre modèle vocal unique.</p>
                    </div>
                ) : (
                    <>
                        <p className="text-lg font-serif italic text-[#1A1A1A]/80 mb-8 px-4 leading-relaxed">
                            "Bonjour, je vous appelle concernant la facture en attente. Nous aimerions régulariser cela rapidement."
                        </p>

                        <div className="flex justify-center mb-8">
                            {isRecording ? (
                                <div className="flex items-center gap-1 h-12">
                                    {[...Array(5)].map((_, i) => (
                                        <motion.div
                                            key={i}
                                            animate={{ height: [10, 40, 10] }}
                                            transition={{ repeat: Infinity, duration: 0.5, delay: i * 0.1 }}
                                            className="w-1.5 bg-red-500 rounded-full"
                                        />
                                    ))}
                                </div>
                            ) : (
                                <div className="h-12 flex items-center text-[#1A1A1A]/20 font-bold text-xs uppercase tracking-widest">
                                    Prêt à enregistrer
                                </div>
                            )}
                        </div>

                        <button
                            onClick={handleRecord}
                            className={cn(
                                "w-20 h-20 rounded-full flex items-center justify-center shadow-xl transition-all mx-auto",
                                isRecording ? "bg-red-500 text-white scale-110" : "bg-[#1A1A1A] text-white hover:bg-[#1A4D2E]"
                            )}
                        >
                            {isRecording ? <StopCircle size={32} fill="currentColor" /> : <Mic size={32} />}
                        </button>
                        <p className="mt-4 text-xs font-bold text-[#1A1A1A]/30 uppercase tracking-widest">
                            {isRecording ? "Appuyez pour terminer" : "Appuyez pour parler"}
                        </p>
                    </>
                )}
            </div>

            <button
                onClick={onNext} // Skip for demo
                className="w-full py-3 rounded-xl border border-[#1A1A1A]/10 text-[#1A1A1A]/40 font-bold tracking-wide uppercase text-xs hover:text-[#1A1A1A] transition-colors"
            >
                Passer cette étape
            </button>
        </motion.div >
    );
};

export default function OnboardingPage() {
    const router = useRouter();
    const { logout, user, token, login } = useAuth();
    const [step, setStep] = useState(1);

    const finishOnboarding = async () => {
        if (!user?.cabinet_id || !token) return;
        try {
            await fetch(`http://localhost:8080/api/v1/cabinets/${user.cabinet_id}`, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ onboarding_completed: true })
            });
            // Update local context
            if (user) {
                login(token, { ...user, onboarding_completed: true });
            }
        } catch (err) {
            console.error(err);
            router.push('/dashboard');
        }
    };

    const nextStep = () => {
        if (step < 4) {
            setStep(step + 1);
        } else {
            finishOnboarding();
        }
    };

    return (
        <div className="min-h-screen bg-[#F9F8F6] text-[#1A1A1A] font-sans selection:bg-[#4F2830]/20 flex items-center justify-center p-4">
            <div className="w-full max-w-lg">
                {/* Header */}
                <div className="flex justify-between items-center mb-8">
                    <div className="flex items-center gap-2">
                        <div className="w-8 h-8 bg-[#1A4D2E] rounded-lg flex items-center justify-center text-white">
                            <span className="font-serif font-bold italic">F</span>
                        </div>
                        <span className="font-serif font-bold text-xl">Fiducia</span>
                    </div>
                    <button onClick={logout} className="text-xs font-bold text-[#1A1A1A]/40 hover:text-[#1A1A1A] uppercase tracking-wider">
                        Déconnexion
                    </button>
                </div>

                <StepIndicator currentStep={step} totalSteps={4} />

                <AnimatePresence mode="wait">
                    {step === 1 && <IdentityStep key="step1" onNext={nextStep} />}
                    {step === 2 && <ClientsStep key="step2" onNext={nextStep} />}
                    {step === 3 && <DataStep key="step3" onNext={nextStep} />}
                    {step === 4 && <VoiceStep key="step4" onNext={nextStep} />}
                </AnimatePresence>
            </div>
        </div>
    );
}
