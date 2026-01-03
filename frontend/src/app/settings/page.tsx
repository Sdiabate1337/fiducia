'use client';

import { useState, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    Mic, Upload, Play, Square, Trash2, ArrowLeft,
    CheckCircle2, AlertTriangle, Fingerprint,
    Lock, StopCircle, Settings, Bell, Megaphone,
    User, ChevronRight, Clock, ToggleLeft, ToggleRight
} from 'lucide-react';
import { cn } from '@/lib/utils';
import Link from 'next/link';

// --- Sub-Components (Views) ---

const GeneralSettings = () => (
    <motion.div
        initial={{ opacity: 0, x: 20 }}
        animate={{ opacity: 1, x: 0 }}
        className="space-y-8"
    >
        <div>
            <h2 className="text-2xl font-serif font-bold text-[#1A1A1A]">Général</h2>
            <p className="text-[#1A1A1A]/60">Informations de votre cabinet.</p>
        </div>

        <div className="bg-white rounded-2xl border border-[#1A1A1A]/5 p-6 space-y-4">
            <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                    <label className="text-xs font-bold text-[#1A1A1A]/40 uppercase tracking-widest">Nom du Cabinet</label>
                    <input
                        disabled
                        value="Cabinet Fiducia Demo"
                        className="w-full bg-[#F9F8F6] border-none rounded-xl px-4 py-3 font-medium text-[#1A1A1A]/60"
                    />
                </div>
                <div className="space-y-2">
                    <label className="text-xs font-bold text-[#1A1A1A]/40 uppercase tracking-widest">SIRET</label>
                    <input
                        disabled
                        value="882 123 456 00012"
                        className="w-full bg-[#F9F8F6] border-none rounded-xl px-4 py-3 font-medium text-[#1A1A1A]/60"
                    />
                </div>
            </div>
        </div>
    </motion.div>
);

const CampaignSettings = () => {
    const [quietHoursEnabled, setQuietHoursEnabled] = useState(true);
    const [startHour, setStartHour] = useState("19:00");
    const [endHour, setEndHour] = useState("08:00");

    return (
        <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            className="space-y-8"
        >
            <div>
                <h2 className="text-2xl font-serif font-bold text-[#1A1A1A]">Campagnes</h2>
                <p className="text-[#1A1A1A]/60">Configuration des relances automatiques.</p>
            </div>

            {/* Active Sequence Card */}
            <div className="bg-white rounded-2xl border border-[#1A1A1A]/5 p-6 relative overflow-hidden group">
                <div className="absolute top-0 right-0 p-20 bg-[#1A4D2E]/5 rounded-full blur-3xl group-hover:bg-[#1A4D2E]/10 transition-colors" />

                <div className="flex flex-col md:flex-row md:items-start justify-between relative z-10 gap-6 md:gap-0">
                    <div>
                        <div className="flex items-center gap-2 mb-2">
                            <Megaphone size={20} className="text-[#1A4D2E]" />
                            <h3 className="font-bold text-[#1A1A1A]">Séquence Active</h3>
                        </div>
                        <h4 className="text-xl font-serif mb-1">Smart Sequence v2</h4>
                        <p className="text-sm text-[#1A1A1A]/60 mb-4">
                            4 étapes • J+0, J+2, J+7, J+15
                        </p>
                    </div>
                    <Link
                        href="/campaigns"
                        className="px-4 py-2 bg-[#1A1A1A] text-white rounded-xl text-sm font-medium hover:bg-[#1A4D2E] transition-colors flex items-center justify-center gap-2 w-full md:w-auto"
                    >
                        <Settings size={16} />
                        Modifier la séquence
                    </Link>
                </div>

                {/* Mini Visualization */}
                <div className="flex flex-wrap items-center gap-2 mt-4 relative z-10">
                    {['Email + WhatsApp', 'Relance Vocale', 'Email Sévère', 'Huissier'].map((step, i) => (
                        <div key={i} className="flex items-center gap-2">
                            <div className="px-3 py-1.5 bg-[#F9F8F6] border border-[#1A1A1A]/5 rounded-lg text-xs font-medium text-[#1A1A1A]/60 whitespace-nowrap">
                                {step}
                            </div>
                            {i < 3 && (
                                <>
                                    <div className="w-4 h-px bg-[#1A1A1A]/10 hidden md:block" />
                                    <div className="text-[#1A1A1A]/10 md:hidden">→</div>
                                </>
                            )}
                        </div>
                    ))}
                </div>
            </div>

            {/* Quiet Hours */}
            <div className="bg-white rounded-2xl border border-[#1A1A1A]/5 p-6">
                <div className="flex items-start justify-between mb-6">
                    <div>
                        <div className="flex items-center gap-2 mb-1">
                            <Clock size={20} className="text-[#1A4D2E]" />
                            <h3 className="font-bold text-[#1A1A1A]">Heures Silencieuses</h3>
                        </div>
                        <p className="text-sm text-[#1A1A1A]/60 max-w-sm">
                            Empêcher l'envoi de messages (WhatsApp, Email) en dehors des heures de bureau pour respecter la vie privée.
                        </p>
                    </div>
                    <button onClick={() => setQuietHoursEnabled(!quietHoursEnabled)}>
                        {quietHoursEnabled ? (
                            <ToggleRight size={32} className="text-[#1A4D2E]" />
                        ) : (
                            <ToggleLeft size={32} className="text-[#1A1A1A]/20" />
                        )}
                    </button>
                </div>

                <div className={`grid grid-cols-1 md:grid-cols-2 gap-4 transition-opacity ${quietHoursEnabled ? 'opacity-100' : 'opacity-40 pointer-events-none'}`}>
                    <div className="space-y-2">
                        <label className="text-xs font-bold text-[#1A1A1A]/40 uppercase tracking-widest">Début (Soir)</label>
                        <input
                            type="time"
                            value={startHour}
                            onChange={(e) => setStartHour(e.target.value)}
                            className="w-full bg-[#F9F8F6] border-none rounded-xl px-4 py-3 font-medium text-[#1A1A1A]"
                        />
                    </div>
                    <div className="space-y-2">
                        <label className="text-xs font-bold text-[#1A1A1A]/40 uppercase tracking-widest">Fin (Matin)</label>
                        <input
                            type="time"
                            value={endHour}
                            onChange={(e) => setEndHour(e.target.value)}
                            className="w-full bg-[#F9F8F6] border-none rounded-xl px-4 py-3 font-medium text-[#1A1A1A]"
                        />
                    </div>
                </div>
            </div>
        </motion.div>
    );
};

// Existing Voice Logic Wrapped in Component
const VoiceSettings = () => {
    const [isRecording, setIsRecording] = useState(false);
    const [audioBlob, setAudioBlob] = useState<Blob | null>(null);
    const [audioUrl, setAudioUrl] = useState<string | null>(null);
    const [voiceName, setVoiceName] = useState('');
    const [uploading, setUploading] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

    const mediaRecorderRef = useRef<MediaRecorder | null>(null);
    const chunksRef = useRef<Blob[]>([]);
    const fileInputRef = useRef<HTMLInputElement>(null);
    const audioRef = useRef<HTMLAudioElement>(null);
    const [isPlaying, setIsPlaying] = useState(false);

    const startRecording = async () => {
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            const mediaRecorder = new MediaRecorder(stream);
            mediaRecorderRef.current = mediaRecorder;
            chunksRef.current = [];

            mediaRecorder.ondataavailable = (e) => {
                if (e.data.size > 0) chunksRef.current.push(e.data);
            };

            mediaRecorder.onstop = () => {
                const blob = new Blob(chunksRef.current, { type: 'audio/webm' });
                setAudioBlob(blob);
                setAudioUrl(URL.createObjectURL(blob));
                stream.getTracks().forEach(track => track.stop());
            };

            mediaRecorder.start();
            setIsRecording(true);
        } catch (err) {
            setMessage({ type: 'error', text: 'Accès micro refusé' });
        }
    };

    const stopRecording = () => {
        if (mediaRecorderRef.current && isRecording) {
            mediaRecorderRef.current.stop();
            setIsRecording(false);
        }
    };

    const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) {
            setAudioBlob(file);
            setAudioUrl(URL.createObjectURL(file));
        }
    };

    const clearAudio = () => {
        setAudioBlob(null);
        setAudioUrl(null);
        if (fileInputRef.current) fileInputRef.current.value = '';
    };

    const togglePlay = () => {
        if (audioRef.current) {
            if (isPlaying) {
                audioRef.current.pause();
            } else {
                audioRef.current.play();
            }
            setIsPlaying(!isPlaying);
        }
    };

    const uploadVoice = async () => {
        if (!audioBlob || !voiceName.trim()) return;

        setUploading(true);
        setMessage(null);

        try {
            const formData = new FormData();
            formData.append('audio', audioBlob, 'voice_sample.webm');
            formData.append('name', voiceName);
            const collaboratorId = '22222222-2222-2222-2222-222222222222';

            const res = await fetch(`/api/v1/collaborators/${collaboratorId}/voice/clone`, {
                method: 'POST',
                body: formData,
            });

            if (res.ok) {
                const data = await res.json();
                setMessage({ type: 'success', text: `Voix "${data.name}" clonée avec succès !` });
                clearAudio();
                setVoiceName('');
            } else {
                setMessage({ type: 'error', text: 'Échec du clonage' });
            }
        } catch (err) {
            setMessage({ type: 'error', text: 'Erreur réseau' });
        } finally {
            setUploading(false);
        }
    };

    return (
        <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            className="space-y-8"
        >
            <div>
                <h2 className="text-2xl font-serif font-bold text-[#1A1A1A]">Voice AI</h2>
                <p className="text-[#1A1A1A]/60">Calibrage de votre clone vocal.</p>
            </div>

            <div className="bg-white/80 backdrop-blur-xl rounded-[2rem] shadow-xl shadow-[#1A1A1A]/5 border border-[#1A1A1A]/5 overflow-hidden">
                <AnimatePresence>
                    {message && (
                        <motion.div
                            initial={{ opacity: 0, height: 0 }}
                            animate={{ opacity: 1, height: 'auto' }}
                            exit={{ opacity: 0, height: 0 }}
                            className={cn(
                                "px-6 py-3 flex items-center justify-center gap-2 text-sm font-medium",
                                message.type === 'success' ? "bg-[#1A4D2E]/10 text-[#1A4D2E]" : "bg-red-50 text-red-600"
                            )}
                        >
                            {message.type === 'success' ? <CheckCircle2 size={16} /> : <AlertTriangle size={16} />}
                            {message.text}
                        </motion.div>
                    )}
                </AnimatePresence>

                <div className="p-8 md:p-12">
                    <div className="h-48 mb-10 relative flex items-center justify-center">
                        <div className="absolute inset-0 flex items-center justify-center">
                            <div className="w-32 h-32 rounded-full border border-[#1A1A1A]/5" />
                            <div className="w-20 h-20 rounded-full border border-[#1A1A1A]/10 absolute" />
                        </div>
                        {isRecording ? (
                            <div className="flex items-center gap-1.5 h-full z-10">
                                {[...Array(8)].map((_, i) => (
                                    <motion.div
                                        key={i}
                                        animate={{ height: [20, Math.random() * 60 + 20, 20] }}
                                        transition={{ repeat: Infinity, duration: 0.4, delay: i * 0.05 }}
                                        className="w-1.5 bg-[#1A1A1A] rounded-full"
                                    />
                                ))}
                            </div>
                        ) : audioUrl ? (
                            <div className="relative z-10 text-center">
                                <button onClick={togglePlay} className="w-16 h-16 bg-[#1A1A1A] rounded-full flex items-center justify-center text-white mb-4 shadow-lg mx-auto">
                                    {isPlaying ? <Square size={20} fill="currentColor" /> : <Play size={24} fill="currentColor" className="ml-1" />}
                                </button>
                                <audio ref={audioRef} src={audioUrl} onEnded={() => setIsPlaying(false)} className="hidden" />
                                <button onClick={clearAudio} className="text-xs uppercase font-bold tracking-widest text-red-500 hover:text-red-600 flex items-center gap-2 mx-auto">
                                    <Trash2 size={12} /> Supprimer
                                </button>
                            </div>
                        ) : (
                            <div className="relative z-10 text-center text-[#1A1A1A]/30">
                                <Mic size={32} className="mx-auto mb-2" strokeWidth={1} />
                                <p className="font-serif text-[#1A1A1A]/60">Prêt</p>
                            </div>
                        )}
                    </div>

                    <div className="flex flex-col gap-6">
                        {!audioUrl ? (
                            <div className="flex flex-col gap-4">
                                {isRecording ? (
                                    <button onClick={stopRecording} className="w-full py-4 rounded-xl bg-red-500 text-white font-bold tracking-wide uppercase shadow-lg flex items-center justify-center gap-2">
                                        <StopCircle size={18} fill="currentColor" /> Arrêter
                                    </button>
                                ) : (
                                    <div className="flex gap-4">
                                        <button onClick={startRecording} className="flex-1 py-4 rounded-xl bg-[#1A1A1A] text-white font-bold tracking-wide uppercase shadow-lg flex items-center justify-center gap-2">
                                            <Mic size={18} /> Enregistrer
                                        </button>
                                        <button onClick={() => fileInputRef.current?.click()} className="px-4 py-4 rounded-xl bg-white border border-[#1A1A1A]/10 hover:bg-[#F9F8F6] flex items-center justify-center">
                                            <Upload size={18} />
                                            <input ref={fileInputRef} type="file" hidden accept="audio/*" onChange={handleFileUpload} />
                                        </button>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <div className="space-y-4">
                                <input
                                    value={voiceName}
                                    onChange={(e) => setVoiceName(e.target.value)}
                                    placeholder="Nom du clone..."
                                    className="w-full bg-[#F9F8F6] border-none rounded-xl px-4 py-3 font-serif outline-none"
                                />
                                <button onClick={uploadVoice} disabled={!voiceName.trim() || uploading} className="w-full py-4 rounded-xl bg-[#1A4D2E] text-white font-bold tracking-wide uppercase shadow-lg disabled:opacity-50 flex items-center justify-center gap-2">
                                    {uploading ? "Clonage..." : <><Fingerprint size={18} /> Générer</>}
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </motion.div>
    );
};

// --- Main Layout ---

export default function SettingsPage() {
    const [activeTab, setActiveTab] = useState<'general' | 'campaigns' | 'voice'>('voice');

    const menuItems = [
        { id: 'general', label: 'Général', icon: Settings },
        { id: 'campaigns', label: 'Campagnes', icon: Megaphone },
        { id: 'voice', label: 'Voice AI', icon: Mic },
        { id: 'notifications', label: 'Notifications', icon: Bell, disabled: true },
        { id: 'users', label: 'Utilisateurs', icon: User, disabled: true },
    ];

    return (
        <div className="min-h-screen bg-[#F9F8F6] text-[#1A1A1A] font-sans selection:bg-[#4F2830]/20 flex">
            {/* Sidebar */}
            <aside className="w-64 border-r border-[#1A1A1A]/5 bg-white/50 backdrop-blur-md hidden md:block fixed h-full z-20">
                <div className="p-8">
                    <Link href="/dashboard" className="flex items-center gap-2 text-[#1A1A1A]/40 hover:text-[#1A1A1A] transition-colors mb-8 group">
                        <ArrowLeft size={16} className="group-hover:-translate-x-1 transition-transform" />
                        <span className="text-sm font-medium">Retour</span>
                    </Link>
                    <h1 className="text-xl font-serif font-bold text-[#1A1A1A] mb-8">Paramètres</h1>

                    <nav className="space-y-1">
                        {menuItems.map((item) => (
                            <button
                                key={item.id}
                                disabled={item.disabled}
                                onClick={() => setActiveTab(item.id as any)}
                                className={cn(
                                    "w-full flex items-center justify-between px-4 py-3 rounded-xl text-sm font-medium transition-all",
                                    activeTab === item.id
                                        ? "bg-[#1A1A1A] text-white shadow-lg"
                                        : "text-[#1A1A1A]/60 hover:bg-[#1A1A1A]/5",
                                    item.disabled && "opacity-40 cursor-not-allowed"
                                )}
                            >
                                <div className="flex items-center gap-3">
                                    <item.icon size={18} />
                                    {item.label}
                                </div>
                                {activeTab === item.id && <ChevronRight size={14} />}
                            </button>
                        ))}
                    </nav>
                </div>
            </aside>


            {/* Mobile Header */}
            <div className="md:hidden fixed top-0 w-full bg-white/80 backdrop-blur border-b border-[#1A1A1A]/5 z-20 px-4 h-16 flex items-center justify-between">
                <Link href="/dashboard"><ArrowLeft size={20} /></Link>
                <span className="font-serif font-bold">Paramètres</span>
                <div className="w-5" />
            </div>

            {/* Mobile Bottom Navigation */}
            <div className="md:hidden fixed bottom-0 left-0 w-full bg-white border-t border-[#1A1A1A]/5 z-30 pb-safe">
                <div className="flex justify-around items-center h-16 px-2">
                    {menuItems.filter(item => !item.disabled).map((item) => (
                        <button
                            key={item.id}
                            onClick={() => setActiveTab(item.id as any)}
                            className={cn(
                                "flex flex-col items-center justify-center gap-1 p-2 rounded-xl transition-all w-20",
                                activeTab === item.id
                                    ? "text-[#1A1A1A]"
                                    : "text-[#1A1A1A]/40"
                            )}
                        >
                            <item.icon size={20} strokeWidth={activeTab === item.id ? 2.5 : 1.5} />
                            <span className="text-[10px] font-medium">{item.label}</span>
                        </button>
                    ))}
                </div>
            </div>

            {/* Content Area */}
            <main className="flex-1 md:ml-64 p-4 md:p-12 pt-20 pb-24 md:pt-12 md:pb-12 max-w-4xl">
                {activeTab === 'general' && <GeneralSettings />}
                {activeTab === 'campaigns' && <CampaignSettings />}
                {activeTab === 'voice' && <VoiceSettings />}
            </main>
        </div>
    );
}


