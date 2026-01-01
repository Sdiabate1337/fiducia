'use client';

import { useState, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    Mic, Upload, Play, Square, Trash2, ArrowLeft,
    CheckCircle2, AlertTriangle, Speaker, Fingerprint,
    Waveform, Lock, Clock, StopCircle
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useRouter } from 'next/navigation';
import Link from 'next/link';

export default function SettingsPage() {
    const router = useRouter();
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
        <div className="min-h-screen bg-[#F9F8F6] text-[#1A1A1A] font-sans selection:bg-[#4F2830]/20 flex flex-col items-center justify-center p-4 md:p-8 relative overflow-hidden">

            {/* Ambient Background */}
            <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-[#1A4D2E]/5 rounded-full blur-3xl" />

            <div className="w-full max-w-xl relative z-10">
                {/* Header with Navigation */}
                <div className="flex items-center justify-between mb-8">
                    <Link href="/dashboard" className="p-2 -ml-2 text-[#1A1A1A]/40 hover:text-[#1A1A1A] transition-colors rounded-lg hover:bg-[#1A1A1A]/5">
                        <ArrowLeft size={24} />
                    </Link>
                    <div className="text-center">
                        <h1 className="text-2xl font-serif font-bold text-[#1A1A1A]">Voice Cloning Lab</h1>
                        <p className="text-xs font-bold uppercase tracking-widest text-[#1A1A1A]/40 mt-1">Calibrage Biométrique</p>
                    </div>
                    <div className="w-10" />
                </div>

                {/* Main Card */}
                <div className="bg-white/80 backdrop-blur-xl rounded-[2rem] shadow-2xl shadow-[#1A1A1A]/5 border border-[#1A1A1A]/5 overflow-hidden">

                    {/* Status Message */}
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

                        {/* Visualization Area */}
                        <div className="h-64 mb-10 relative flex items-center justify-center">
                            {/* Static Rings */}
                            <div className="absolute inset-0 flex items-center justify-center">
                                <div className="w-48 h-48 rounded-full border border-[#1A1A1A]/5" />
                                <div className="w-32 h-32 rounded-full border border-[#1A1A1A]/10 absolute" />
                            </div>

                            {/* Dynamic Content */}
                            {isRecording ? (
                                <div className="flex items-center gap-1.5 h-full z-10">
                                    {[...Array(12)].map((_, i) => (
                                        <motion.div
                                            key={i}
                                            animate={{
                                                height: [20, Math.random() * 80 + 30, 20],
                                            }}
                                            transition={{
                                                repeat: Infinity,
                                                duration: 0.4,
                                                delay: i * 0.05
                                            }}
                                            className="w-2 bg-[#1A1A1A] rounded-full"
                                        />
                                    ))}
                                </div>
                            ) : audioUrl ? (
                                <div className="relative z-10 text-center">
                                    <button
                                        onClick={togglePlay}
                                        className="w-20 h-20 bg-[#1A1A1A] rounded-full flex items-center justify-center text-white mb-6 hover:scale-105 transition-transform shadow-xl mx-auto"
                                    >
                                        {isPlaying ? <Square size={24} fill="currentColor" /> : <Play size={28} fill="currentColor" className="ml-1" />}
                                    </button>
                                    <audio
                                        ref={audioRef}
                                        src={audioUrl}
                                        onEnded={() => setIsPlaying(false)}
                                        className="hidden"
                                    />
                                    <button
                                        onClick={clearAudio}
                                        className="text-xs uppercase font-bold tracking-widest text-red-500 hover:text-red-600 flex items-center gap-2 mx-auto"
                                    >
                                        <Trash2 size={14} /> Supprimer l'échantillon
                                    </button>
                                </div>
                            ) : (
                                <div className="relative z-10 text-center text-[#1A1A1A]/30">
                                    <Mic size={48} className="mx-auto mb-4" strokeWidth={1} />
                                    <p className="font-serif text-lg text-[#1A1A1A]/60">En attente d'échantillon</p>
                                </div>
                            )}
                        </div>

                        {/* Controls */}
                        <div className="flex flex-col gap-6">
                            {!audioUrl ? (
                                <div className="flex flex-col gap-4">
                                    {isRecording ? (
                                        <button
                                            onClick={stopRecording}
                                            className="w-full py-5 rounded-2xl bg-red-500 text-white font-bold tracking-wide uppercase transition-all shadow-lg shadow-red-500/30 hover:shadow-red-500/40 hover:-translate-y-0.5 flex items-center justify-center gap-3"
                                        >
                                            <StopCircle size={20} fill="currentColor" /> Arrêter l'enregistrement
                                        </button>
                                    ) : (
                                        <div className="flex gap-4">
                                            <button
                                                onClick={startRecording}
                                                className="flex-1 py-5 rounded-2xl bg-[#1A1A1A] text-white font-bold tracking-wide uppercase transition-all shadow-lg hover:shadow-xl hover:-translate-y-0.5 flex items-center justify-center gap-3"
                                            >
                                                <Mic size={20} /> Enregistrer
                                            </button>
                                            <button
                                                onClick={() => fileInputRef.current?.click()}
                                                className="px-6 py-5 rounded-2xl bg-white border border-[#1A1A1A]/10 text-[#1A1A1A] font-medium transition-all hover:bg-[#F9F8F6] flex items-center justify-center"
                                            >
                                                <Upload size={20} />
                                                <input
                                                    ref={fileInputRef}
                                                    type="file"
                                                    hidden
                                                    accept="audio/*"
                                                    onChange={handleFileUpload}
                                                />
                                            </button>
                                        </div>
                                    )}
                                    <p className="text-center text-xs text-[#1A1A1A]/40 mt-2">
                                        Durée recommandée : 30 à 60 secondes de parole naturelle.
                                    </p>
                                </div>
                            ) : (
                                <motion.div
                                    initial={{ opacity: 0, y: 20 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    className="space-y-6"
                                >
                                    <div className="space-y-2">
                                        <label className="text-xs font-bold text-[#1A1A1A]/40 uppercase tracking-widest">Nom du Clone</label>
                                        <input
                                            value={voiceName}
                                            onChange={(e) => setVoiceName(e.target.value)}
                                            placeholder="ex: Jean Dupont - Expert"
                                            className="w-full bg-[#F9F8F6] border-none rounded-xl px-4 py-4 font-serif text-lg outline-none focus:ring-2 focus:ring-[#1A1A1A]/10 placeholder:text-[#1A1A1A]/20"
                                        />
                                    </div>
                                    <button
                                        onClick={uploadVoice}
                                        disabled={!voiceName.trim() || uploading}
                                        className="w-full py-5 rounded-2xl bg-[#1A4D2E] text-white font-bold tracking-wide uppercase transition-all shadow-lg shadow-[#1A4D2E]/20 hover:shadow-[#1A4D2E]/40 hover:-translate-y-0.5 disabled:opacity-50 disabled:hover:translate-y-0 flex items-center justify-center gap-3"
                                    >
                                        {uploading ? (
                                            <>
                                                <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                                Clonage en cours...
                                            </>
                                        ) : (
                                            <>
                                                <Fingerprint size={20} /> Générer l'empreinte vocale
                                            </>
                                        )}
                                    </button>
                                </motion.div>
                            )}
                        </div>

                    </div>

                    {/* Secure Footer */}
                    <div className="bg-[#F9F8F6] px-8 py-4 border-t border-[#1A1A1A]/5 flex items-center justify-center gap-2 text-xs text-[#1A1A1A]/40 font-medium">
                        <Lock size={12} />
                        Sécurisé par VoiceShield™ • Données cryptées
                    </div>
                </div>
            </div>
        </div>
    );
}
