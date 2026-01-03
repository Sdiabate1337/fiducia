'use client';

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ArrowRight, Lock, CheckCircle2, ShieldCheck, Mail, User } from 'lucide-react';
import Link from 'next/link';
import { useAuth } from '@/context/AuthContext';
import { useRouter } from 'next/navigation';

export default function LoginPage() {
    const { login } = useAuth();
    const router = useRouter();

    // Mode toggler: 'login' | 'register'
    const [mode, setMode] = useState<'login' | 'register'>('login');

    // Form State
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [fullName, setFullName] = useState('');
    const [cabinetName, setCabinetName] = useState('');

    const [status, setStatus] = useState<'idle' | 'loading' | 'error' | 'success'>('idle');
    const [errorMessage, setErrorMessage] = useState('');
    const [successMessage, setSuccessMessage] = useState('');

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setStatus('loading');
        setErrorMessage('');

        try {
            const endpoint = mode === 'login' ? '/api/v1/auth/login' : '/api/v1/auth/register';
            const body = mode === 'login'
                ? { email, password }
                : { email, password, full_name: fullName, cabinet_name: cabinetName };

            const res = await fetch(`http://localhost:8080${endpoint}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body),
            });

            const data = await res.json();

            if (!res.ok) {
                throw new Error(data.error || 'Une erreur est survenue');
            }

            if (mode === 'register') {
                // Success Registration -> Switch to Login
                setStatus('success');
                setSuccessMessage('Compte créé avec succès. Veuillez vous connecter.');
                setMode('login');
                setPassword('');
            } else {
                // Success Login
                login(data.token, data.user);
                // login() handles redirect to dashboard
            }

        } catch (err: any) {
            setStatus('error');
            setErrorMessage(err.message);
        } finally {
            if (status !== 'error') setStatus('idle');
        }
    };

    return (
        <div className="min-h-screen bg-[#F9F8F6] text-[#1A1A1A] font-sans selection:bg-[#4F2830]/20 flex flex-col items-center justify-center p-4 relative overflow-hidden">

            {/* Background Texture/Gradient */}
            <div className="absolute inset-0 z-0">
                <div className="absolute top-0 left-0 w-full h-[50vh] bg-gradient-to-b from-[#1A4D2E]/5 to-transparent" />
                <div className="absolute bottom-0 right-0 p-[20vw] bg-[#4F2830]/5 rounded-full blur-3xl translate-x-1/2 translate-y-1/2" />
            </div>

            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.8, ease: "easeOut" }}
                className="w-full max-w-md relative z-10"
            >
                {/* Brand Header */}
                <div className="text-center mb-8">
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        transition={{ delay: 0.2 }}
                        className="inline-flex items-center gap-2 mb-4 px-4 py-2 rounded-full bg-[#1A4D2E]/5 border border-[#1A4D2E]/10"
                    >
                        <ShieldCheck size={14} className="text-[#1A4D2E]" />
                        <span className="text-xs font-bold tracking-widest uppercase text-[#1A4D2E]">Espace Souverain</span>
                    </motion.div>
                    <h1 className="font-serif text-4xl md:text-5xl font-bold text-[#1A1A1A] mb-4 tracking-tight">
                        Fiducia
                    </h1>
                    <p className="text-[#1A1A1A]/60 font-serif text-lg italic leading-relaxed">
                        Zéro compte d'attente.<br />
                        Zéro impayé.
                    </p>
                </div>

                {/* Login Card */}
                <div className="bg-white/80 backdrop-blur-xl rounded-[2rem] shadow-2xl shadow-[#1A1A1A]/10 border border-[#1A1A1A]/5 p-8 md:p-10 relative overflow-hidden">
                    {/* Top Accent Line */}
                    <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-[#1A4D2E] via-[#4F2830] to-[#1A4D2E]" />

                    <form onSubmit={handleSubmit} className="space-y-5">
                        <AnimatePresence mode="popLayout">
                            {/* REGISTER FIELDS */}
                            {mode === 'register' && (
                                <motion.div
                                    key="register-fields"
                                    initial={{ opacity: 0, height: 0 }}
                                    animate={{ opacity: 1, height: 'auto' }}
                                    exit={{ opacity: 0, height: 0 }}
                                    className="space-y-5 overflow-hidden"
                                >
                                    <div className="space-y-1">
                                        <label className="block text-xs font-bold uppercase tracking-widest text-[#1A1A1A]/40 ml-1">Nom du Cabinet</label>
                                        <input
                                            required
                                            value={cabinetName}
                                            onChange={(e) => setCabinetName(e.target.value)}
                                            placeholder="Cabinet Expert..."
                                            className="w-full bg-[#F9F8F6] border border-[#1A1A1A]/5 rounded-xl px-5 py-3 font-medium text-[#1A1A1A] outline-none focus:ring-2 focus:ring-[#1A4D2E]/20"
                                        />
                                    </div>
                                    <div className="space-y-1">
                                        <label className="block text-xs font-bold uppercase tracking-widest text-[#1A1A1A]/40 ml-1">Votre Nom</label>
                                        <div className="relative">
                                            <input
                                                required
                                                value={fullName}
                                                onChange={(e) => setFullName(e.target.value)}
                                                placeholder="Jean Dupont"
                                                className="w-full bg-[#F9F8F6] border border-[#1A1A1A]/5 rounded-xl px-5 py-3 font-medium text-[#1A1A1A] outline-none focus:ring-2 focus:ring-[#1A4D2E]/20"
                                            />
                                            <User size={18} className="absolute right-4 top-1/2 -translate-y-1/2 text-[#1A1A1A]/20" />
                                        </div>
                                    </div>
                                </motion.div>
                            )}

                            {/* COMMON FIELDS */}
                            <motion.div layout key="common-fields" className="space-y-5">
                                <div className="space-y-1">
                                    <label className="block text-xs font-bold uppercase tracking-widest text-[#1A1A1A]/40 ml-1">Email Professionnel</label>
                                    <div className="relative group">
                                        <input
                                            type="email"
                                            required
                                            value={email}
                                            onChange={(e) => setEmail(e.target.value)}
                                            placeholder="nom@cabinet.com"
                                            className="w-full bg-[#F9F8F6] border border-[#1A1A1A]/5 rounded-xl px-5 py-4 font-medium text-[#1A1A1A] placeholder:text-[#1A1A1A]/20 outline-none focus:ring-2 focus:ring-[#1A4D2E]/20 focus:border-[#1A4D2E]/30 transition-all group-hover:bg-[#F9F8F6]/80"
                                        />
                                        <div className="absolute right-4 top-1/2 -translate-y-1/2 text-[#1A1A1A]/20">
                                            <Mail size={18} />
                                        </div>
                                    </div>
                                </div>

                                <div className="space-y-1">
                                    <label className="block text-xs font-bold uppercase tracking-widest text-[#1A1A1A]/40 ml-1">Mot de Passe</label>
                                    <div className="relative group">
                                        <input
                                            type="password"
                                            required
                                            value={password}
                                            onChange={(e) => setPassword(e.target.value)}
                                            placeholder="••••••••"
                                            className="w-full bg-[#F9F8F6] border border-[#1A1A1A]/5 rounded-xl px-5 py-4 font-medium text-[#1A1A1A] placeholder:text-[#1A1A1A]/20 outline-none focus:ring-2 focus:ring-[#1A4D2E]/20 focus:border-[#1A4D2E]/30 transition-all group-hover:bg-[#F9F8F6]/80"
                                        />
                                        <div className="absolute right-4 top-1/2 -translate-y-1/2 text-[#1A1A1A]/20">
                                            <Lock size={18} />
                                        </div>
                                    </div>
                                </div>
                            </motion.div>
                        </AnimatePresence>

                        {/* Success Message */}
                        {successMessage && (
                            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="p-3 bg-green-50 text-green-700 text-sm rounded-lg text-center font-medium flex items-center justify-center gap-2">
                                <CheckCircle2 size={16} />
                                {successMessage}
                            </motion.div>
                        )}

                        {/* Error Message */}
                        {status === 'error' && (
                            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="p-3 bg-red-50 text-red-600 text-sm rounded-lg text-center font-medium">
                                {errorMessage}
                            </motion.div>
                        )}

                        <button
                            type="submit"
                            disabled={status === 'loading'}
                            className="w-full py-4 rounded-xl bg-[#1A1A1A] text-white font-bold tracking-wide uppercase shadow-lg shadow-[#1A1A1A]/20 hover:shadow-xl hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50 disabled:pointer-events-none flex items-center justify-center gap-3"
                        >
                            {status === 'loading' ? (
                                <>
                                    <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                    Traitement...
                                </>
                            ) : (
                                <>
                                    {mode === 'login' ? 'Connexion Sécurisée' : 'Créer mon Espace'} <ArrowRight size={18} />
                                </>
                            )}
                        </button>
                    </form>

                    <div className="mt-6 text-center pt-6 border-t border-[#1A1A1A]/5">
                        <p className="text-sm text-[#1A1A1A]/40">
                            {mode === 'login' ? "Pas encore membre ?" : "Déjà un compte ?"}
                            <button
                                onClick={() => setMode(mode === 'login' ? 'register' : 'login')}
                                className="ml-2 font-bold text-[#1A1A1A] hover:underline"
                            >
                                {mode === 'login' ? "Demander une invitation" : "Se connecter"}
                            </button>
                        </p>
                    </div>

                </div>
            </motion.div>
        </div>
    );
}
