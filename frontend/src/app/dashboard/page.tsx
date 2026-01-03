'use client';

import { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    Search, Filter, Receipt, Upload, Settings,
    AlertCircle, CheckCircle2, Clock, Send, XCircle,
    ArrowUpRight, MoreHorizontal, ChevronDown, Download, LogOut, ArrowRight
} from 'lucide-react';
import Link from 'next/link';
import { useAuth } from '@/context/AuthContext';
import { useRouter } from 'next/navigation';

// --- Types ---

interface PendingLine {
    id: string;
    amount: string;
    transaction_date: string;
    bank_label: string | null;
    status: string;
    client?: {
        id: string;
        name: string;
        phone: string | null;
    };
    campaign_status?: string;
    campaign_current_step?: number;
    next_step_scheduled_at?: string;
}

interface Stats {
    total: number;
    pending: number;
    contacted: number;
    received: number;
    validated: number;
    rejected: number;
    pending_amount: number;
    validated_amount: number;
}

interface ListResponse {
    items: PendingLine[];
    total: number;
    has_more: boolean;
}

// --- Components ---

const StatCard = ({ label, value, icon: Icon, color, isActive, onClick }: any) => (
    <div
        onClick={onClick}
        className={`p-6 rounded-2xl border transition-all duration-300 cursor-pointer group ${isActive
            ? 'bg-[#1A1A1A] text-white border-[#1A1A1A]'
            : 'bg-white border-[#1A1A1A]/5 hover:border-[#1A1A1A]/20 hover:shadow-lg'
            }`}
    >
        <div className="flex justify-between items-start mb-4">
            <div className={`p-2 rounded-lg ${isActive ? 'bg-white/10' : 'bg-[#F9F8F6]'}`}>
                <Icon size={20} className={isActive ? 'text-white' : color} />
            </div>
            {isActive && <div className="h-2 w-2 rounded-full bg-[#20E070]" />}
        </div>
        <div className="space-y-1">
            <h3 className={`text-sm font-medium ${isActive ? 'text-white/60' : 'text-[#1A1A1A]/60'}`}>{label}</h3>
            <p className={`text-3xl font-serif ${isActive ? 'text-white' : 'text-[#1A1A1A]'}`}>{value}</p>
        </div>
    </div>
);

const VelocityCard = ({ stats, formatAmount }: { stats: Stats, formatAmount: (a: string) => string }) => (
    <div className="col-span-1 md:col-span-2 bg-[#1A1A1A] rounded-2xl p-6 md:p-8 text-white relative overflow-hidden">
        {/* Background Pattern */}
        <div className="absolute top-0 right-0 w-64 h-64 bg-[#1A4D2E] blur-[100px] opacity-20 transform translate-x-1/2 -translate-y-1/2" />

        <div className="relative z-10 flex flex-col justify-between h-full">
            <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4 mb-8">
                <div>
                    <h3 className="text-white/60 text-xs md:text-sm font-medium uppercase tracking-widest mb-2">Cash Velocity</h3>
                    <div className="text-3xl md:text-5xl font-serif">
                        {formatAmount((stats.validated_amount + stats.pending_amount).toString())}
                    </div>
                </div>
                <div className="md:text-right">
                    <span className="inline-flex items-center px-3 py-1 bg-[#20E070]/20 text-[#20E070] text-[10px] md:text-xs font-bold uppercase tracking-wider rounded-full">
                        +12% vs M-1
                    </span>
                </div>
            </div>

            <div className="grid grid-cols-2 gap-4 md:gap-8 border-t border-white/10 pt-6">
                <div>
                    <div className="text-white/40 text-[10px] md:text-xs uppercase tracking-wider mb-1">En Attente</div>
                    <div className="text-xl md:text-2xl font-serif text-[#F59E0B]">{formatAmount(stats.pending_amount.toString())}</div>
                </div>
                <div>
                    <div className="text-white/40 text-[10px] md:text-xs uppercase tracking-wider mb-1">Sécurisé</div>
                    <div className="text-xl md:text-2xl font-serif text-[#20E070]">{formatAmount(stats.validated_amount.toString())}</div>
                </div>
            </div>
        </div>
    </div>
);

interface OnboardingStatus {
    has_clients: boolean;
    client_count: number;
    has_lines: boolean;
    line_count: number;
    voice_configured: boolean;
}

const SetupWrapper = ({ status, onClose }: { status: OnboardingStatus, onClose: () => void }) => {
    // Determine progress
    const steps = [
        { key: 'clients', label: "Importer le carnet d'adresses", done: status.has_clients, link: '/onboarding?step=2' },
        { key: 'lines', label: "Importer le Grand Livre (471)", done: status.has_lines, link: '/import' },
        { key: 'voice', label: "Calibrer sa voix IA", done: status.voice_configured, link: '/settings?tab=voice' },
    ];
    const progress = steps.filter(s => s.done).length;
    const total = steps.length;
    const isComplete = progress === total;

    if (isComplete) return null;

    return (
        <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95 }}
            className="mb-8 bg-white rounded-2xl border border-[#1A1A1A]/10 p-6 shadow-xl shadow-[#1A1A1A]/5 relative overflow-hidden"
        >
            <div className="absolute top-0 left-0 w-1 h-full bg-[#F59E0B]" />
            <div className="flex flex-col md:flex-row justify-between md:items-center gap-6">
                <div>
                    <h3 className="text-xl font-serif font-bold text-[#1A1A1A] mb-2">Bienvenue sur Fiducia</h3>
                    <p className="text-[#1A1A1A]/60 max-w-lg">
                        Votre espace n'est pas encore totalement configuré. Suivez ces étapes pour activer le pilotage automatique.
                    </p>
                    <div className="mt-4 flex items-center gap-2">
                        <div className="h-2 w-32 bg-[#1A1A1A]/10 rounded-full overflow-hidden">
                            <div className="h-full bg-[#F59E0B]" style={{ width: `${(progress / total) * 100}%` }} />
                        </div>
                        <span className="text-xs font-bold text-[#1A1A1A]/60">{progress}/{total} étapes</span>
                    </div>
                </div>

                <div className="grid gap-3">
                    {steps.map((step, i) => (
                        <Link
                            key={step.key}
                            href={step.link}
                            className={`flex items-center gap-3 p-3 rounded-xl border transition-all ${step.done
                                ? 'bg-[#1A4D2E]/5 border-[#1A4D2E]/10 text-[#1A4D2E]'
                                : 'bg-white border-[#1A1A1A]/10 text-[#1A1A1A] hover:border-[#1A1A1A]/30 hover:shadow-sm'
                                }`}
                        >
                            <div className={`w-6 h-6 rounded-full flex items-center justify-center border ${step.done ? 'bg-[#1A4D2E] border-[#1A4D2E] text-white' : 'border-[#1A1A1A]/20 text-[#1A1A1A]/40'
                                }`}>
                                {step.done ? <CheckCircle2 size={14} /> : <span className="text-xs font-bold">{i + 1}</span>}
                            </div>
                            <span className={`text-sm font-medium ${step.done ? 'line-through opacity-60' : ''}`}>{step.label}</span>
                            {!step.done && <ArrowRight size={14} className="ml-auto opacity-40" >
                                <ArrowRight size={14} />
                            </ArrowRight>}
                        </Link>
                    ))}
                </div>
            </div>

            <button
                onClick={onClose}
                className="absolute top-4 right-4 text-[#1A1A1A]/20 hover:text-[#1A1A1A] transition-colors"
            >
                <XCircle size={20} />
            </button>
        </motion.div>
    );
};

// --- Page ---

export default function DashboardPage() {
    const { logout, user, token } = useAuth();
    const [lines, setLines] = useState<PendingLine[]>([]);
    const [stats, setStats] = useState<Stats | null>(null);
    const [loading, setLoading] = useState(true);
    const [statusFilter, setStatusFilter] = useState<string>('all');
    const [search, setSearch] = useState('');
    const [isFilterMenuOpen, setIsFilterMenuOpen] = useState(false);
    const [isProfileMenuOpen, setIsProfileMenuOpen] = useState(false);
    const [onboardingStatus, setOnboardingStatus] = useState<OnboardingStatus | null>(null);
    const [showSetup, setShowSetup] = useState(true);

    const toggleFilterMenu = () => setIsFilterMenuOpen(!isFilterMenuOpen);
    const toggleProfileMenu = () => setIsProfileMenuOpen(!isProfileMenuOpen);

    const selectFilter = (status: string) => {
        setStatusFilter(status);
        setIsFilterMenuOpen(false);
    };

    // Demo cabinet ID or real one
    const cabinetId = user?.cabinet_id || '00000000-0000-0000-0000-000000000001';

    useEffect(() => {
        if (token) fetchData();
    }, [statusFilter, search, token]);

    const fetchData = async () => {
        setLoading(true);
        try {
            const statsRes = await fetch(`/api/v1/cabinets/${cabinetId}/pending-lines/stats`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (statsRes.ok) {
                const data = await statsRes.json();
                setStats(data);
            }

            // Fetch Onboarding Status
            const statusRes = await fetch(`/api/v1/cabinets/${cabinetId}/onboarding-status`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (statusRes.ok) {
                const statusData = await statusRes.json();
                setOnboardingStatus(statusData);
            }

            let url = `/api/v1/cabinets/${cabinetId}/pending-lines?limit=50`;
            if (statusFilter !== 'all') url += `&status=${statusFilter}`;
            if (search) url += `&search=${encodeURIComponent(search)}`;

            const linesRes = await fetch(url, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (linesRes.ok) {
                const linesData: ListResponse = await linesRes.json();
                let items = linesData.items || [];

                // Fallback: Client-side filtering in case backend is stale
                if (statusFilter !== 'all') {
                    items = items.filter(line => line.status === statusFilter);
                }

                setLines(items);
            }
        } catch (err) {
            console.error('Failed to fetch data:', err);
        } finally {
            setLoading(false);
        }
    };

    const formatAmount = (amount: string) => {
        const num = parseFloat(amount);
        return new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'MAD' }).format(num);
    };

    const formatDate = (dateStr: string) => {
        return new Date(dateStr).toLocaleDateString('fr-FR', { day: '2-digit', month: 'short' });
    };

    const StatusBadge = ({ status, mobile = false }: { status: string, mobile?: boolean }) => {
        const styles: any = {
            pending: { color: 'bg-amber-100 text-amber-800', label: 'En attente', icon: Clock },
            contacted: { color: 'bg-blue-100 text-blue-800', label: 'Contacté', icon: Send },
            received: { color: 'bg-purple-100 text-purple-800', label: 'Reçu', icon: Receipt },
            validated: { color: 'bg-green-100 text-green-800', label: 'Validé', icon: CheckCircle2 },
            rejected: { color: 'bg-red-100 text-red-800', label: 'Rejeté', icon: XCircle },
        };
        const style = styles[status] || styles.pending;
        const Icon = style.icon;

        return (
            <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${style.color} ${mobile ? 'text-[10px] px-2 py-0.5' : ''}`}>
                <Icon size={mobile ? 10 : 12} />
                {style.label}
            </span>
        );
    };

    return (
        <div className="min-h-screen bg-[#F9F8F6] text-[#1A1A1A] font-sans selection:bg-[#4F2830]/20">
            {/* Navbar */}
            <nav className="sticky top-0 z-50 bg-[#F9F8F6]/80 backdrop-blur-md border-b border-[#1A1A1A]/5 px-4 md:px-8 h-16 md:h-20 flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <Link href="/" className="text-xl md:text-2xl font-serif font-bold tracking-tight">Fiducia.</Link>
                    <div className="h-6 w-px bg-[#1A1A1A]/10 mx-2 hidden md:block" />
                    <span className="text-sm font-medium text-[#1A1A1A]/50 hidden md:block">Cabinet Demo</span>
                </div>
                <div className="flex items-center gap-2 md:gap-4">
                    <button className="p-2 text-[#1A1A1A]/40 hover:text-[#1A1A1A] transition-colors md:hidden">
                        <Search size={20} />
                    </button>
                    <button className="p-2 text-[#1A1A1A]/40 hover:text-[#1A1A1A] transition-colors hidden md:block">
                        <Search size={20} />
                    </button>
                    <Link href="/settings" className="p-2 text-[#1A1A1A]/40 hover:text-[#1A1A1A] transition-colors">
                        <Settings size={20} />
                    </Link>
                    <div className="relative">
                        <button
                            onClick={toggleProfileMenu}
                            className="w-8 h-8 rounded-full bg-[#1A1A1A] text-white flex items-center justify-center font-serif text-sm hover:ring-2 hover:ring-[#1A1A1A]/20 transition-all"
                        >
                            {user?.full_name ? user.full_name.charAt(0).toUpperCase() : 'JD'}
                        </button>

                        <AnimatePresence>
                            {isProfileMenuOpen && (
                                <motion.div
                                    initial={{ opacity: 0, y: 10, scale: 0.95 }}
                                    animate={{ opacity: 1, y: 0, scale: 1 }}
                                    exit={{ opacity: 0, y: 10, scale: 0.95 }}
                                    className="absolute top-full right-0 mt-2 w-48 bg-white rounded-xl shadow-xl border border-[#1A1A1A]/5 z-50 overflow-hidden origin-top-right"
                                >
                                    <div className="p-1">
                                        <div className="px-4 py-3 border-b border-[#1A1A1A]/5 mb-1">
                                            <p className="text-sm font-bold text-[#1A1A1A] truncate">{user?.full_name}</p>
                                            <p className="text-xs text-[#1A1A1A]/40 truncate">{user?.email}</p>
                                        </div>
                                        <button
                                            onClick={logout}
                                            className="w-full text-left px-4 py-2 text-sm font-medium text-red-600 hover:bg-red-50 rounded-lg transition-colors flex items-center gap-2"
                                        >
                                            <LogOut size={14} />
                                            Déconnexion
                                        </button>
                                    </div>
                                </motion.div>
                            )}
                        </AnimatePresence>
                    </div>
                </div>
            </nav>

            <main className="max-w-7xl mx-auto px-4 md:px-6 py-8 md:py-12">
                <div className="flex flex-col md:flex-row justify-between items-start md:items-end mb-8 md:mb-12 gap-4">
                    <div>
                        <h1 className="text-3xl md:text-4xl font-serif mb-2 text-[#1A1A1A]">Vue d'ensemble</h1>
                        <p className="text-[#1A1A1A]/60 text-sm md:text-base">Gérez vos flux et pilotez la performance du cabinet.</p>
                    </div>
                    <Link href="/import" className="w-full md:w-auto px-6 py-3 bg-[#1A1A1A] text-white rounded-xl font-medium hover:bg-[#1A4D2E] transition-all flex items-center justify-center gap-2 shadow-lg hover:shadow-xl">
                        <Upload size={18} />
                        Importer un relevé
                    </Link>
                </div>

                {/* Setup Guide */}
                <AnimatePresence>
                    {showSetup && onboardingStatus && (
                        <SetupWrapper status={onboardingStatus} onClose={() => setShowSetup(false)} />
                    )}
                </AnimatePresence>

                {/* Stats Grid */}
                {stats && (
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4 md:gap-6 mb-8 md:mb-12">
                        <VelocityCard stats={stats} formatAmount={formatAmount} />

                        <div className="md:col-span-2 grid grid-cols-2 gap-3 md:gap-6">
                            <StatCard
                                label="En attente"
                                value={stats.pending}
                                icon={Clock}
                                color="text-amber-600"
                                isActive={statusFilter === 'pending'}
                                onClick={() => setStatusFilter(statusFilter === 'pending' ? 'all' : 'pending')}
                            />
                            <StatCard
                                label="Sécurisés"
                                value={stats.validated}
                                icon={CheckCircle2}
                                color="text-[#20E070]"
                                isActive={statusFilter === 'validated'}
                                onClick={() => setStatusFilter(statusFilter === 'validated' ? 'all' : 'validated')}
                            />
                            <StatCard
                                label="Contactés"
                                value={stats.contacted}
                                icon={Send}
                                color="text-blue-600"
                                isActive={statusFilter === 'contacted'}
                                onClick={() => setStatusFilter(statusFilter === 'contacted' ? 'all' : 'contacted')}
                            />
                            <StatCard
                                label="Reçus"
                                value={stats.received}
                                icon={Receipt}
                                color="text-purple-600"
                                isActive={statusFilter === 'received'}
                                onClick={() => setStatusFilter(statusFilter === 'received' ? 'all' : 'received')}
                            />
                        </div>
                    </div>
                )}

                {/* Filters Row */}
                <div className="flex flex-col md:flex-row gap-4 md:gap-6 mb-8">
                    <div className="relative flex-1 group">
                        <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 text-[#1A1A1A]/40 group-focus-within:text-[#1A1A1A] transition-colors" size={20} />
                        <input
                            className="w-full pl-12 pr-4 py-3 md:py-4 bg-white border border-[#1A1A1A]/5 rounded-xl outline-none focus:border-[#1A1A1A]/20 transition-all font-medium text-[#1A1A1A] shadow-sm text-sm md:text-base"
                            placeholder="Rechercher une transaction, un montant..."
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                        />
                    </div>
                    <div className="flex flex-wrap gap-2 md:gap-3 pb-2 md:pb-0 relative z-20">
                        <div className="relative">
                            <button
                                onClick={toggleFilterMenu}
                                className={`px-4 md:px-6 py-3 md:py-4 bg-white border rounded-xl font-medium transition-all flex items-center gap-2 whitespace-nowrap text-sm md:text-base shadow-sm ${isFilterMenuOpen ? 'border-[#1A1A1A] bg-[#1A1A1A] text-white' : 'border-[#1A1A1A]/5 text-[#1A1A1A]/70 hover:border-[#1A1A1A]/20 hover:text-[#1A1A1A]'}`}
                            >
                                <Filter size={18} />
                                Filtres
                                <ChevronDown size={14} className={`transition-transform duration-200 ${isFilterMenuOpen ? 'rotate-180' : ''}`} />
                            </button>

                            <AnimatePresence>
                                {isFilterMenuOpen && (
                                    <motion.div
                                        initial={{ opacity: 0, y: 10, scale: 0.95 }}
                                        animate={{ opacity: 1, y: 0, scale: 1 }}
                                        exit={{ opacity: 0, y: 10, scale: 0.95 }}
                                        className="absolute top-full right-0 mt-2 w-56 bg-white rounded-xl shadow-xl border border-[#1A1A1A]/5 z-50 overflow-hidden origin-top-right"
                                    >
                                        <div className="p-1">
                                            {[
                                                { label: 'Tout voir', value: 'all' },
                                                { label: 'En attente', value: 'pending' },
                                                { label: 'Contacté', value: 'contacted' },
                                                { label: 'Reçu (à valider)', value: 'received' },
                                                { label: 'Validé', value: 'validated' },
                                                { label: 'Rejeté', value: 'rejected' },
                                            ].map((option) => (
                                                <button
                                                    key={option.value}
                                                    onClick={() => selectFilter(option.value)}
                                                    className={`w-full text-left px-4 py-3 text-sm font-medium rounded-lg transition-colors flex items-center justify-between ${statusFilter === option.value ? 'bg-[#F9F8F6] text-[#1A1A1A]' : 'text-[#1A1A1A]/60 hover:bg-[#F9F8F6] hover:text-[#1A1A1A]'}`}
                                                >
                                                    {option.label}
                                                    {statusFilter === option.value && <CheckCircle2 size={16} className="text-[#1A4D2E]" />}
                                                </button>
                                            ))}
                                        </div>
                                    </motion.div>
                                )}
                            </AnimatePresence>
                        </div>
                        {statusFilter !== 'all' && (
                            <button
                                onClick={() => setStatusFilter('all')}
                                className="px-4 md:px-6 py-3 md:py-4 bg-[#F9F8F6] text-[#1A1A1A]/60 hover:text-red-600 font-medium transition-colors whitespace-nowrap text-sm md:text-base"
                            >
                                Effacer
                            </button>
                        )}
                    </div>
                </div>

                {/* Data List (Card View for Mobile, Table for Desktop) */}
                <div className="bg-white rounded-2xl border border-[#1A1A1A]/5 shadow-xl shadow-[#1A1A1A]/5 overflow-hidden">
                    {loading ? (
                        <div className="p-20 text-center">
                            <div className="animate-spin w-8 h-8 border-2 border-[#1A1A1A] border-t-transparent rounded-full mx-auto mb-4" />
                            <p className="text-[#1A1A1A]/40 font-medium animate-pulse">Synchronisation bancaire...</p>
                        </div>
                    ) : lines.length === 0 ? (
                        <div className="p-20 text-center">
                            <div className="w-16 h-16 bg-[#F9F8F6] rounded-full flex items-center justify-center mx-auto mb-6">
                                <Search className="text-[#1A1A1A]/20" size={32} />
                            </div>
                            <h3 className="text-lg font-serif mb-2">Aucun résultat</h3>
                            <p className="text-[#1A1A1A]/40 max-w-sm mx-auto">Aucune transaction ne correspond à vos critères de recherche.</p>
                        </div>
                    ) : (
                        <>
                            {/* Mobile Card View */}
                            <div className="block md:hidden">
                                {lines.map((line, i) => (
                                    <motion.div
                                        key={line.id}
                                        initial={{ opacity: 0, y: 10 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        transition={{ delay: i * 0.05 }}
                                        className="p-4 border-b border-[#1A1A1A]/5 last:border-none"
                                    >
                                        <div className="flex justify-between items-start mb-3">
                                            <div>
                                                <div className="text-sm font-semibold text-[#1A1A1A] mb-1">{line.bank_label}</div>
                                                <div className="text-xs text-[#1A1A1A]/40 flex items-center gap-1">
                                                    <span>{formatDate(line.transaction_date)}</span>
                                                    <span>•</span>
                                                    <span>ID: ...{line.id.slice(-4)}</span>
                                                </div>
                                            </div>
                                            <div className="text-right">
                                                <div className="text-sm font-serif font-bold text-[#1A1A1A] mb-1">
                                                    {formatAmount(line.amount)}
                                                </div>
                                                <StatusBadge status={line.status} mobile={true} />
                                            </div>
                                        </div>
                                        <div className="flex justify-between items-center">
                                            {line.client ? (
                                                <div className="flex items-center gap-2">
                                                    <div className="w-6 h-6 rounded-full bg-[#1A1A1A]/5 flex items-center justify-center text-[10px] font-bold text-[#1A1A1A]/60">
                                                        {line.client.name.slice(0, 2).toUpperCase()}
                                                    </div>
                                                    <span className="text-xs text-[#1A1A1A]/70 font-medium">{line.client.name}</span>
                                                </div>
                                            ) : (
                                                <span className="text-xs text-[#1A1A1A]/30 italic px-2 py-1 rounded bg-[#1A1A1A]/5">Non assigné</span>
                                            )}

                                            <Link
                                                href={`/pending-lines/${line.id}`}
                                                className="px-4 py-2 bg-[#F9F8F6] border border-[#1A1A1A]/5 rounded-lg text-xs font-medium text-[#1A1A1A] hover:bg-[#1A1A1A] hover:text-white transition-all"
                                            >
                                                Gérer
                                            </Link>
                                        </div>
                                    </motion.div>
                                ))}
                            </div>

                            {/* Desktop Table View */}
                            <div className="hidden md:block overflow-x-auto">
                                <table className="w-full">
                                    <thead className="bg-[#F9F8F6] border-b border-[#1A1A1A]/5">
                                        <tr>
                                            <th className="px-8 py-5 text-left text-xs font-semibold text-[#1A1A1A]/40 uppercase tracking-wider">Date</th>
                                            <th className="px-8 py-5 text-left text-xs font-semibold text-[#1A1A1A]/40 uppercase tracking-wider">Libellé Bancaire</th>
                                            <th className="px-8 py-5 text-left text-xs font-semibold text-[#1A1A1A]/40 uppercase tracking-wider">Client</th>
                                            <th className="px-8 py-5 text-left text-xs font-semibold text-[#1A1A1A]/40 uppercase tracking-wider">Montant</th>
                                            <th className="px-8 py-5 text-left text-xs font-semibold text-[#1A1A1A]/40 uppercase tracking-wider">Statut</th>
                                            <th className="px-8 py-5 text-left text-xs font-semibold text-[#1A1A1A]/40 uppercase tracking-wider">Relance</th>
                                            <th className="px-8 py-5 text-right text-xs font-semibold text-[#1A1A1A]/40 uppercase tracking-wider">Action</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-[#1A1A1A]/5">
                                        {lines.map((line, i) => (
                                            <motion.tr
                                                key={line.id}
                                                initial={{ opacity: 0, y: 10 }}
                                                animate={{ opacity: 1, y: 0 }}
                                                transition={{ delay: i * 0.05 }}
                                                className="group hover:bg-[#F9F8F6]/50 transition-colors"
                                            >
                                                <td className="px-8 py-6 text-sm text-[#1A1A1A]/60 font-medium">
                                                    {formatDate(line.transaction_date)}
                                                </td>
                                                <td className="px-8 py-6">
                                                    <div className="text-sm font-medium text-[#1A1A1A] max-w-xs truncate">{line.bank_label}</div>
                                                    <div className="text-xs text-[#1A1A1A]/40 mt-1">ID: #...{line.id.slice(-4)}</div>
                                                </td>
                                                <td className="px-8 py-6">
                                                    {line.client ? (
                                                        <div className="flex items-center gap-3">
                                                            <div className="w-8 h-8 rounded-full bg-[#1A1A1A]/5 flex items-center justify-center text-xs font-bold text-[#1A1A1A]/60">
                                                                {line.client.name.slice(0, 2).toUpperCase()}
                                                            </div>
                                                            <div className="text-sm font-medium text-[#1A1A1A]/80">{line.client.name}</div>
                                                        </div>
                                                    ) : (
                                                        <span className="text-xs text-[#1A1A1A]/30 italic px-2 py-1 rounded bg-[#1A1A1A]/5">Non assigné</span>
                                                    )}
                                                </td>
                                                <td className="px-8 py-6 text-sm font-bold text-[#1A1A1A] font-serif">
                                                    {formatAmount(line.amount)}
                                                </td>
                                                <td className="px-8 py-6">
                                                    <StatusBadge status={line.status} />
                                                </td>
                                                <td className="px-8 py-6">
                                                    {line.campaign_status ? (
                                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${line.campaign_status === 'running' ? 'bg-blue-100 text-blue-800' :
                                                            line.campaign_status === 'stopped' ? 'bg-red-100 text-red-800' :
                                                                line.campaign_status === 'completed' ? 'bg-green-100 text-green-800' :
                                                                    'bg-gray-100 text-gray-800'
                                                            }`}>
                                                            {line.campaign_status === 'running' ? `En cours (Etape ${line.campaign_current_step || 1})` :
                                                                line.campaign_status === 'stopped' ? 'Stoppé' :
                                                                    line.campaign_status === 'completed' ? 'Terminé' :
                                                                        'En attente'}
                                                        </span>
                                                    ) : (
                                                        <span className="text-gray-300">-</span>
                                                    )}
                                                </td>
                                                <td className="px-8 py-6 text-right">
                                                    <Link
                                                        href={`/pending-lines/${line.id}`}
                                                        className="inline-flex items-center justify-center px-4 py-2 border border-[#1A1A1A]/10 rounded-lg text-sm font-medium text-[#1A1A1A]/70 hover:bg-[#1A1A1A] hover:text-white transition-all hover:shadow-md"
                                                    >
                                                        Gérer
                                                    </Link>
                                                </td>
                                            </motion.tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </>
                    )}
                </div>
            </main>
        </div>
    );
}
