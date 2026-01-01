'use client';

import { motion } from 'framer-motion';
import { ArrowRight, CheckCircle2, BarChart3, ShieldCheck, Zap, Info } from 'lucide-react';
import { useState } from 'react';
import Link from 'next/link';

// Custom Minimal components for Landing Page to match Editorial Style
const FadeIn = ({ children, delay = 0 }: { children: React.ReactNode; delay?: number }) => (
    <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.8, delay, ease: [0.22, 1, 0.36, 1] }}
    >
        {children}
    </motion.div>
);

export default function LandingPage() {
    const [billing, setBilling] = useState<'monthly' | 'annual'>('monthly');

    const prices = {
        discovery: billing === 'monthly' ? 590 : 472,
        growth: billing === 'monthly' ? 1690 : 1352,
        elite: billing === 'monthly' ? 3490 : 2792
    }

    return (
        <div className="min-h-screen bg-[#F9F8F6] text-[#1A1A1A] font-sans selection:bg-[#4F2830]/20">
            {/* Navbar */}
            <nav className="fixed top-0 w-full z-50 bg-[#F9F8F6]/80 backdrop-blur-md border-b border-[#1A1A1A]/5">
                <div className="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
                    <div className="flex items-center gap-12">
                        <Link href="/" className="text-2xl font-serif font-bold tracking-tight">Fiducia.</Link>
                        <div className="hidden md:flex gap-8 text-sm font-medium text-[#1A1A1A]/70">
                            <Link href="#features" className="hover:text-[#1A1A1A] transition-colors">Plateforme</Link>
                            <Link href="#solutions" className="hover:text-[#1A1A1A] transition-colors">Solutions</Link>
                            <Link href="#pricing" className="hover:text-[#1A1A1A] transition-colors">Tarifs</Link>
                        </div>
                    </div>
                    <div className="flex items-center gap-4">
                        <Link href="/login" className="text-sm font-medium hover:text-[#4F2830] transition-colors">Connexion</Link>
                        <Link href="/dashboard" className="bg-[#1A1A1A] text-white px-5 py-2.5 rounded-full text-sm font-medium hover:bg-[#4F2830] transition-colors">
                            Démarrer
                        </Link>
                    </div>
                </div>
            </nav>

            {/* Hero Section */}
            <section className="pt-40 pb-20 md:pt-48 md:pb-32 px-6">
                <div className="max-w-7xl mx-auto grid md:grid-cols-2 gap-16 items-center">
                    <FadeIn>
                        <h1 className="text-5xl md:text-7xl font-serif font-medium leading-[1.1] mb-8 text-[#1A1A1A]">
                            Zéro compte d'attente. <br />
                            <span className="italic text-[#1A4D2E]">Zéro impayé.</span>
                        </h1>
                        <p className="text-lg md:text-xl text-[#1A1A1A]/60 mb-10 max-w-md leading-relaxed">
                            Ne courez plus après les justificatifs. Fiducia automatise la collecte, la relance vocale et la validation comptable. Gagnez 15 heures de production par collaborateur, chaque mois.
                        </p>
                        <div className="flex flex-col sm:flex-row gap-4">
                            <Link href="/dashboard" className="inline-flex items-center justify-center px-8 py-4 bg-[#1A1A1A] text-white rounded-full font-medium hover:bg-[#1A4D2E] transition-all transform hover:-translate-y-1">
                                Explorer la Démo <ArrowRight className="ml-2 h-4 w-4" />
                            </Link>
                            <button className="inline-flex items-center justify-center px-8 py-4 bg-white border border-[#1A1A1A]/10 text-[#1A1A1A] rounded-full font-medium hover:bg-white hover:border-[#1A1A1A]/30 transition-all">
                                Voir la vidéo
                            </button>
                        </div>

                        <div className="mt-12 flex items-center gap-4 text-sm text-[#1A1A1A]/50">
                            <div className="flex -space-x-2">
                                {[1, 2, 3, 4].map(i => (
                                    <div key={i} className="w-8 h-8 rounded-full bg-gray-200 border-2 border-[#F9F8F6]" />
                                ))}
                            </div>
                            <p>Déjà utilisé par +500 cabinets modernes</p>
                        </div>
                    </FadeIn>

                    <FadeIn delay={0.2}>
                        <div className="relative">
                            <div className="absolute inset-0 bg-[#A3C4BC]/20 rounded-[2rem] transform rotate-3 scale-105" />
                            <div className="relative bg-white rounded-[2rem] shadow-2xl overflow-hidden border border-[#1A1A1A]/5 aspect-[4/5] md:aspect-square">
                                {/* Abstract UI Representation */}
                                <div className="p-8 h-full flex flex-col">
                                    <div className="flex justify-between items-center mb-8">
                                        <div className="w-12 h-12 rounded-full bg-[#F9F8F6] flex items-center justify-center">
                                            <BarChart3 className="w-6 h-6 text-[#1A1A1A]" />
                                        </div>
                                        <div className="px-3 py-1 bg-[#20E070]/10 text-[#1A4D2E] rounded-full text-xs font-medium">
                                            +24% Cash Flow
                                        </div>
                                    </div>
                                    <div className="space-y-4 flex-1">
                                        {[1, 2, 3].map((i) => (
                                            <div key={i} className="p-4 rounded-xl bg-[#F9F8F6] border border-[#1A1A1A]/5 flex items-center gap-4">
                                                <div className="w-10 h-10 rounded-full bg-white flex items-center justify-center text-lg">✨</div>
                                                <div className="flex-1">
                                                    <div className="h-2 w-24 bg-[#1A1A1A]/10 rounded mb-2" />
                                                    <div className="h-2 w-16 bg-[#1A1A1A]/5 rounded" />
                                                </div>
                                                <div className="text-sm font-serif font-medium">2,450 €</div>
                                            </div>
                                        ))}

                                        <div className="mt-8 p-6 bg-[#1A4D2E] rounded-2xl text-white relative overflow-hidden">
                                            <div className="relative z-10">
                                                <div className="text-sm opacity-80 mb-1">Status IA</div>
                                                <div className="text-xl font-serif">Autopilot Actif</div>
                                            </div>
                                            <div className="absolute right-0 bottom-0 opacity-10">
                                                <Zap className="w-24 h-24" />
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </FadeIn>
                </div>
            </section>



            {/* Feature 1: Validation */}
            <section id="features" className="py-32 px-6">
                <div className="max-w-7xl mx-auto grid md:grid-cols-2 gap-20 items-center">
                    <FadeIn>
                        <div className="bg-white p-2 rounded-2xl shadow-xl border border-[#1A1A1A]/5">
                            <div className="bg-[#F9F8F6] rounded-xl aspect-[4/3] flex items-center justify-center relative overflow-hidden">
                                <div className="absolute inset-0 flex items-center justify-center text-[#1A1A1A]/10">
                                    <ShieldCheck className="w-32 h-32" />
                                </div>
                                {/* Floating elements */}
                                <motion.div
                                    animate={{ y: [0, -10, 0] }}
                                    transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
                                    className="bg-white p-4 rounded-lg shadow-lg absolute top-1/4 right-1/4 max-w-[200px]"
                                >
                                    <div className="flex items-center gap-2 mb-2">
                                        <div className="w-2 h-2 bg-green-500 rounded-full" />
                                        <div className="text-xs font-medium">Facture Validée</div>
                                    </div>
                                    <div className="h-2 w-full bg-gray-100 rounded" />
                                </motion.div>
                            </div>
                        </div>
                    </FadeIn>
                    <FadeIn delay={0.2}>
                        <div className="inline-block px-3 py-1 bg-[#1A4D2E]/10 text-[#1A4D2E] rounded-full text-xs font-semibold uppercase tracking-wider mb-6">
                            Nettoyage Automatisé du 471
                        </div>
                        <h2 className="text-4xl md:text-5xl font-serif mb-6 text-[#1A1A1A]">
                            Le Compte d'Attente (471) <br />
                            <span className="italic text-[#1A4D2E]">ne sera plus votre goulot.</span>
                        </h2>
                        <p className="text-lg text-[#1A1A1A]/60 mb-8 leading-relaxed">
                            L'IA Fiducia détecte les flux orphelins en temps réel. Elle engage le client sur WhatsApp avec votre voix, récupère la photo du ticket, et pré-remplit l'écriture.
                        </p>
                        <ul className="space-y-4">
                            {[
                                'Détection immédiate des flux orphelins',
                                'Relance vocale humanisée sur WhatsApp',
                                'Transformez 3 semaines de chasse en 48h'
                            ].map((item, i) => (
                                <li key={i} className="flex items-center gap-3">
                                    <CheckCircle2 className="w-5 h-5 text-[#1A4D2E]" />
                                    <span className="text-[#1A1A1A]/80">{item}</span>
                                </li>
                            ))}
                        </ul>
                    </FadeIn>
                </div>
            </section>

            {/* Feature 2: Voice */}
            <section id="solutions" className="py-32 px-6 bg-[#1A1A1A] text-[#F9F8F6] relative overflow-hidden">
                {/* Decorative */}
                <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-[#4F2830] rounded-full blur-[120px] opacity-20" />

                <div className="max-w-7xl mx-auto grid md:grid-cols-2 gap-20 items-center relative z-10">
                    <FadeIn>
                        <div className="inline-block px-3 py-1 bg-white/10 text-white rounded-full text-xs font-semibold uppercase tracking-wider mb-6">
                            L'Inimitable
                        </div>
                        <h2 className="text-4xl md:text-5xl font-serif mb-6 text-white">
                            Votre expertise. Votre voix. <br />
                            <span className="italic opacity-60">À l'échelle.</span>
                        </h2>
                        <p className="text-lg text-white/60 mb-8 leading-relaxed">
                            La relance est une question de dosage. Notre technologie de clonage vocal reproduit votre empathie et votre fermeté. Vos clients reçoivent une note vocale personnelle, pas un rappel froid.
                        </p>
                        <div className="mb-8 flex items-center gap-3 text-white/80">
                            <CheckCircle2 className="w-5 h-5 text-[#20E070]" />
                            <span className="font-medium">85% de taux de réponse en moins de 10 minutes</span>
                        </div>
                        <Link href="/settings" className="text-white border-b border-white pb-1 hover:opacity-70 transition-opacity">
                            Découvrir le Cloning Lab &rarr;
                        </Link>
                    </FadeIn>
                    <FadeIn delay={0.2}>
                        <div className="relative">
                            <div className="absolute inset-0 bg-gradient-to-r from-indigo-500 to-purple-500 rounded-2xl opacity-20 blur-xl" />
                            <div className="bg-white/5 backdrop-blur-lg border border-white/10 p-8 rounded-2xl relative">
                                <div className="flex items-center justify-center h-48">
                                    <div className="flex items-end gap-1 h-24">
                                        {[...Array(20)].map((_, i) => (
                                            <motion.div
                                                key={i}
                                                animate={{
                                                    height: [20, Math.random() * 80 + 20, 20],
                                                    opacity: [0.3, 1, 0.3]
                                                }}
                                                transition={{
                                                    repeat: Infinity,
                                                    duration: 0.8,
                                                    delay: i * 0.05
                                                }}
                                                className="w-2 bg-white rounded-full"
                                            />
                                        ))}
                                    </div>
                                </div>
                                <div className="text-center mt-6">
                                    <div className="text-sm font-medium text-white">Simulation en cours...</div>
                                    <div className="text-xs text-white/40 mt-1">Voix : "Empathie Ferme"</div>
                                </div>
                            </div>
                        </div>
                    </FadeIn>
                </div>
            </section>

            {/* Pricing Section */}
            <section id="pricing" className="py-32 px-6 bg-white relative">
                <div className="absolute inset-0 bg-[radial-gradient(#e5e7eb_1px,transparent_1px)] [background-size:16px_16px] opacity-30" />
                <div className="max-w-7xl mx-auto relative z-10">
                    <div className="text-center mb-16">
                        <h2 className="text-4xl md:text-5xl font-serif mb-6 text-[#1A1A1A]">Plans & Tarifs.</h2>
                        <p className="text-xl md:text-2xl text-[#1A1A1A] max-w-3xl mx-auto mb-10 font-medium leading-relaxed">
                            Gagnez 15 heures de production par collaborateur, chaque mois.
                        </p>

                        {/* Toggle (Visual) */}
                        <div className="inline-flex bg-[#F9F8F6] p-1 rounded-full border border-[#1A1A1A]/5 relative cursor-pointer">
                            <motion.div
                                layout
                                className="absolute top-1 bottom-1 bg-white rounded-full shadow-sm z-0"
                                initial={false}
                                animate={{
                                    left: billing === 'monthly' ? '4px' : '50%',
                                    right: billing === 'monthly' ? '50%' : '4px'
                                }}
                            />
                            <button
                                onClick={() => setBilling('monthly')}
                                className={`px-6 py-2 text-sm font-medium z-10 transition-colors relative w-32 ${billing === 'monthly' ? 'text-[#1A1A1A]' : 'text-[#1A1A1A]/50'}`}
                            >
                                Mensuel
                            </button>
                            <button
                                onClick={() => setBilling('annual')}
                                className={`px-6 py-2 text-sm font-medium z-10 transition-colors relative w-32 ${billing === 'annual' ? 'text-[#1A1A1A]' : 'text-[#1A1A1A]/50'}`}
                            >
                                Annuel <span className="text-xs text-[#1A4D2E] font-bold ml-1">-20%</span>
                            </button>
                        </div>
                    </div>

                    <div className="grid md:grid-cols-3 gap-8 items-start">
                        {/* Plan 1: Découverte */}
                        <FadeIn delay={0.1}>
                            <div className="p-8 rounded-3xl border border-[#1A1A1A]/10 bg-[#F9F8F6] hover:bg-white hover:shadow-xl hover:-translate-y-1 transition-all duration-300">
                                <div className="mb-8">
                                    <h3 className="font-serif text-2xl mb-2 text-[#1A1A1A]">Découverte</h3>
                                    <p className="text-[#1A1A1A]/60 text-sm h-10">L'essentiel pour démarrer la digitalisation.</p>
                                </div>
                                <div className="mb-8 flex items-baseline">
                                    <motion.span
                                        key={prices.discovery}
                                        initial={{ opacity: 0, y: 10 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        className="text-5xl font-bold tracking-tight text-[#1A1A1A]"
                                    >
                                        {prices.discovery}
                                    </motion.span>
                                    <span className="text-lg font-medium ml-2 text-[#1A1A1A]/60">MAD</span>
                                </div>
                                <div className="space-y-4 mb-8">
                                    <div className="text-xs font-bold uppercase tracking-wider text-[#1A1A1A]/40 mb-2">Comprend :</div>
                                    <ul className="space-y-3 text-sm text-[#1A1A1A]/80">
                                        <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#1A4D2E] shrink-0" /> <span>5 Dossiers Inclus</span></li>
                                        <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#1A4D2E] shrink-0" /> <span>WhatsApp <span className="font-semibold">Texte</span> + Email</span></li>
                                        <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#1A4D2E] shrink-0" /> <span>Import Manuel (Excel)</span></li>
                                        <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#1A4D2E] shrink-0" /> <span>Dashboard Basique</span></li>
                                    </ul>
                                </div>
                                <div className="pt-8 border-t border-[#1A1A1A]/5 mt-auto">
                                    <div className="text-xs text-center text-[#1A4D2E] mb-4 font-mono font-semibold">Frais de mise en service offerts</div>
                                    <button className="w-full py-4 rounded-xl border border-[#1A1A1A]/20 font-semibold hover:bg-[#1A1A1A] hover:text-white hover:border-[#1A1A1A] transition-all">
                                        Démarrer
                                    </button>
                                </div>
                            </div>
                        </FadeIn>

                        {/* Plan 2: Croissance - High Viz */}
                        <FadeIn delay={0.2}>
                            <div className="p-1 rounded-3xl bg-gradient-to-b from-[#1A4D2E] to-[#1A1A1A] shadow-2xl relative z-10 transform md:-translate-y-6">
                                <div className="absolute top-0 inset-x-0 h-px bg-white/20" />
                                <div className="bg-[#1A1A1A] rounded-[1.3rem] p-8 h-full flex flex-col">
                                    <div className="mb-2 flex justify-between items-start">
                                        <h3 className="font-serif text-3xl text-white">Croissance</h3>
                                        <span className="bg-[#1A4D2E] text-white text-[10px] font-bold px-3 py-1 rounded-full uppercase tracking-widest border border-white/10 shadow-lg glow">
                                            Recommandé
                                        </span>
                                    </div>
                                    <p className="text-white/60 text-sm mb-8 h-10">L'Agent Virtuel Hybride pour démultiplier la production.</p>

                                    <div className="mb-8 flex items-baseline">
                                        <motion.span
                                            key={prices.growth}
                                            initial={{ opacity: 0, y: 10 }}
                                            animate={{ opacity: 1, y: 0 }}
                                            className="text-6xl font-bold tracking-tight text-white"
                                        >
                                            {prices.growth}
                                        </motion.span>
                                        <span className="text-xl font-medium ml-2 text-white/60">MAD</span>
                                    </div>

                                    <div className="space-y-4 mb-8 flex-1">
                                        <div className="text-xs font-bold uppercase tracking-wider text-white/40 mb-2">Tout Découverte, plus :</div>
                                        <ul className="space-y-3 text-sm text-white font-medium">
                                            <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#20E070] shrink-0 shadow-[0_0_10px_rgba(32,224,112,0.4)]" /> <span>30 Dossiers Inclus</span></li>
                                            <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#20E070] shrink-0" /> <span>WhatsApp <span className="text-[#20E070]">Voice AI</span> + Email</span></li>
                                            <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#20E070] shrink-0" /> <span>Sync Sage / Cegid / Topaze</span></li>
                                            <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#20E070] shrink-0" /> <span>1 Voix Clonée (Collaborateur)</span></li>
                                        </ul>
                                    </div>
                                    <div className="pt-8 border-t border-white/10 text-center">
                                        <div className="text-xs text-white/30 mb-4 font-mono">Setup : 1 500 MAD</div>
                                        <button className="w-full py-4 rounded-xl bg-white text-[#1A1A1A] font-bold text-lg hover:bg-[#F9F8F6] hover:shadow-[0_0_20px_rgba(255,255,255,0.3)] transition-all transform hover:scale-[1.02]">
                                            Choisir l'Agent Virtuel
                                        </button>
                                        <p className="text-[10px] text-white/30 mt-3">Essai gratuit 14 jours, sans CB.</p>
                                    </div>
                                </div>
                            </div>
                        </FadeIn>

                        {/* Plan 3: Cabinet Élite */}
                        <FadeIn delay={0.3}>
                            <div className="p-8 rounded-3xl border border-[#1A1A1A]/10 bg-white hover:border-[#1A1A1A]/30 hover:shadow-xl transition-all duration-300">
                                <div className="mb-8">
                                    <h3 className="font-serif text-2xl mb-2 text-[#1A1A1A]">Cabinet Élite</h3>
                                    <p className="text-[#1A1A1A]/60 text-sm h-10">Pour les cabinets en transformation totale.</p>
                                </div>
                                <div className="mb-8 flex items-baseline">
                                    <motion.span
                                        key={prices.elite}
                                        initial={{ opacity: 0, y: 10 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        className="text-5xl font-bold tracking-tight text-[#1A1A1A]"
                                    >
                                        {prices.elite}
                                    </motion.span>
                                    <span className="text-lg font-medium ml-2 text-[#1A1A1A]/60">MAD</span>
                                </div>
                                <div className="space-y-4 mb-8">
                                    <div className="text-xs font-bold uppercase tracking-wider text-[#1A1A1A]/40 mb-2">Tout Croissance, plus :</div>
                                    <ul className="space-y-3 text-sm text-[#1A1A1A]/80">
                                        <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#4F2830] shrink-0" /> <span>100 Dossiers Inclus</span></li>
                                        <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#4F2830] shrink-0" /> <span>Omni-canal <span className="font-semibold">Illimité</span></span></li>
                                        <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#4F2830] shrink-0" /> <span>Voix Clonées Illimitées</span></li>
                                        <li className="flex items-start gap-3"><CheckCircle2 className="w-5 h-5 text-[#4F2830] shrink-0" /> <span>Manager Dédié (SLA 4h)</span></li>
                                    </ul>
                                </div>
                                <div className="pt-8 border-t border-[#1A1A1A]/5 mt-auto">
                                    <div className="text-xs text-center text-[#1A1A1A]/40 mb-4 font-mono">Setup : 3 000 MAD</div>
                                    <button className="w-full py-4 rounded-xl border border-[#1A1A1A]/20 font-semibold hover:bg-[#1A1A1A] hover:text-white transition-all">
                                        Contact VIP
                                    </button>
                                </div>
                            </div>
                        </FadeIn>
                    </div>
                    <div className="mt-16 text-center">
                        <div className="inline-flex items-center gap-3 px-6 py-4 bg-white border border-[#1A1A1A]/5 rounded-2xl shadow-sm hover:shadow-md transition-shadow">
                            <div className="w-8 h-8 rounded-full bg-[#FFF8E1] flex items-center justify-center text-[#F59E0B]">
                                <Info className="w-5 h-5" />
                            </div>
                            <span className="text-[#1A1A1A]/70 text-sm font-medium">
                                Dossiers supplémentaires facturés entre <span className="text-[#1A1A1A] font-bold">25 et 50 MAD/mois</span> selon volume.
                            </span>
                        </div>
                    </div>
                </div>
            </section>

            {/* Footer */}
            <footer className="bg-[#111] text-white py-20 px-6">
                <div className="max-w-7xl mx-auto grid md:grid-cols-4 gap-12">
                    <div className="col-span-2">
                        <Link href="/" className="text-3xl font-serif font-bold tracking-tight mb-6 block">Fiducia.</Link>
                        <p className="text-white/40 max-w-sm">
                            Le système d&apos;exploitation financier pour les cabinets qui ne veulent plus choisir entre croissance et qualité de vie.
                        </p>
                    </div>
                    <div>
                        <h4 className="font-serif text-lg mb-6">Produit</h4>
                        <ul className="space-y-4 text-white/40 text-sm">
                            <li><Link href="#" className="hover:text-white transition-colors">Fonctionnalités</Link></li>
                            <li><Link href="#" className="hover:text-white transition-colors">Intégrations</Link></li>
                            <li><Link href="#" className="hover:text-white transition-colors">Tarifs</Link></li>
                            <li><Link href="/dashboard" className="hover:text-white transition-colors">Changelog</Link></li>
                        </ul>
                    </div>
                    <div>
                        <h4 className="font-serif text-lg mb-6">Compagnie</h4>
                        <ul className="space-y-4 text-white/40 text-sm">
                            <li><Link href="#" className="hover:text-white transition-colors">À propos</Link></li>
                            <li><Link href="#" className="hover:text-white transition-colors">Carrières</Link></li>
                            <li><Link href="#" className="hover:text-white transition-colors">Blog</Link></li>
                            <li><Link href="#" className="hover:text-white transition-colors">Contact</Link></li>
                        </ul>
                    </div>
                </div>
                <div className="max-w-7xl mx-auto mt-20 pt-8 border-t border-white/10 text-center text-white/20 text-xs">
                    © 2024 Fiducia Inc. Tous droits réservés. Design: Editorial Precision.
                </div>
            </footer>
        </div>
    );
}
