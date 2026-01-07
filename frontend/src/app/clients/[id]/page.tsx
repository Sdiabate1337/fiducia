'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import {
    ArrowLeft, Phone, Mail, Building2, Calendar,
    CreditCard, Clock, CheckCircle2, AlertTriangle,
    FileText, ArrowUpRight, Search, Filter
} from 'lucide-react';
import Link from 'next/link';
import { EditClientModal } from '@/components/clients/EditClientModal';
import { Button } from '@/components/ui/Button';

interface Client {
    id: string;
    name: string;
    phone?: string;
    email?: string;
    siret?: string;
    notes?: string;
    whatsapp_opted_in: boolean;
    created_at: string;
}

interface PendingLine {
    id: string;
    amount: string;
    transaction_date: string;
    bank_label: string;
    status: string;
    created_at: string;
}

const StatusBadge = ({ status }: { status: string }) => {
    const styles: any = {
        pending: { color: 'bg-amber-100 text-amber-800 border-amber-200', label: 'En attente' },
        contacted: { color: 'bg-blue-100 text-blue-800 border-blue-200', label: 'Contacté' },
        received: { color: 'bg-purple-100 text-purple-800 border-purple-200', label: 'Reçu' },
        validated: { color: 'bg-green-100 text-green-800 border-green-200', label: 'Validé' },
        rejected: { color: 'bg-red-100 text-red-800 border-red-200', label: 'Rejeté' },
    };
    const style = styles[status] || styles.pending;
    return (
        <span className={`inline-flex px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider border ${style.color}`}>
            {style.label}
        </span>
    );
};

export default function ClientDetailPage() {
    const { token } = useAuth();
    const params = useParams();
    const router = useRouter();
    const id = params.id as string;

    const [client, setClient] = useState<Client | null>(null);
    const [lines, setLines] = useState<PendingLine[]>([]);
    const [loading, setLoading] = useState(true);
    const [isEditOpen, setIsEditOpen] = useState(false);

    useEffect(() => {
        if (token) fetchData();
    }, [id, token]);

    const fetchData = async () => {
        setLoading(true);
        try {
            const [clientRes, linesRes] = await Promise.all([
                fetch(`http://localhost:8080/api/v1/clients/${id}`, { headers: { Authorization: `Bearer ${token}` } }),
                fetch(`http://localhost:8080/api/v1/clients/${id}/pending-lines`, { headers: { Authorization: `Bearer ${token}` } })
            ]);

            if (clientRes.ok) setClient(await clientRes.json());
            if (linesRes.ok) {
                const data = await linesRes.json();
                setLines(data.items || []);
            }
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const formatAmount = (amount: string) => {
        const num = parseFloat(amount);
        return new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'MAD' }).format(num);
    };

    const formatDate = (dateStr: string) => {
        return new Date(dateStr).toLocaleDateString('fr-FR', { day: '2-digit', month: 'short', year: 'numeric' });
    };

    if (loading) return <div className="h-screen flex items-center justify-center bg-[#F9F8F6]">Chargement...</div>;
    if (!client) return <div className="h-screen flex items-center justify-center bg-[#F9F8F6]">Client introuvable</div>;

    const totalPending = lines.filter(l => l.status === 'pending' || l.status === 'contacted').reduce((acc, curr) => acc + parseFloat(curr.amount), 0);
    const totalValidated = lines.filter(l => l.status === 'validated').reduce((acc, curr) => acc + parseFloat(curr.amount), 0);

    return (
        <div className="min-h-screen bg-[#F9F8F6] text-[#1A1A1A]">
            <EditClientModal
                isOpen={isEditOpen}
                onClose={() => setIsEditOpen(false)}
                client={client}
                onUpdated={(updated) => { setClient(updated); fetchData(); }}
            />

            {/* Header */}
            <div className="bg-white border-b sticky top-0 z-10">
                <div className="max-w-5xl mx-auto px-6 h-16 flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <Button variant="ghost" size="sm" onClick={() => router.back()}>
                            <ArrowLeft className="w-4 h-4 mr-2" /> Retour
                        </Button>
                        <h1 className="font-serif text-xl font-bold">{client.name}</h1>
                    </div>
                    <Button variant="outline" size="sm" onClick={() => setIsEditOpen(true)}>Modifier</Button>
                </div>
            </div>

            <main className="max-w-5xl mx-auto px-6 py-8 grid grid-cols-1 md:grid-cols-3 gap-8">
                {/* Sidebar Info */}
                <div className="space-y-6">
                    <div className="bg-white p-6 rounded-xl border shadow-sm space-y-4">
                        <div className="flex items-center gap-3 text-sm">
                            <div className="w-8 h-8 rounded-full bg-slate-100 flex items-center justify-center">
                                <Building2 className="w-4 h-4 text-slate-500" />
                            </div>
                            <div>
                                <div className="text-muted-foreground text-xs uppercase tracking-wider">SIRET</div>
                                <div className="font-medium">{client.siret || '-'}</div>
                            </div>
                        </div>
                        <div className="flex items-center gap-3 text-sm">
                            <div className="w-8 h-8 rounded-full bg-slate-100 flex items-center justify-center">
                                <Mail className="w-4 h-4 text-slate-500" />
                            </div>
                            <div>
                                <div className="text-muted-foreground text-xs uppercase tracking-wider">Email</div>
                                <div className="font-medium truncate max-w-[200px]">{client.email || '-'}</div>
                            </div>
                        </div>
                        <div className="flex items-center gap-3 text-sm">
                            <div className="w-8 h-8 rounded-full bg-slate-100 flex items-center justify-center">
                                <Phone className="w-4 h-4 text-slate-500" />
                            </div>
                            <div>
                                <div className="text-muted-foreground text-xs uppercase tracking-wider">Téléphone</div>
                                <div className="font-medium">{client.phone || '-'}</div>
                            </div>
                        </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="bg-amber-50 p-4 rounded-xl border border-amber-100">
                            <div className="text-amber-800 text-xs font-bold uppercase tracking-wider mb-1">En suspens</div>
                            <div className="text-xl font-serif font-bold text-amber-900">
                                {new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'MAD', maximumFractionDigits: 0 }).format(totalPending)}
                            </div>
                        </div>
                        <div className="bg-green-50 p-4 rounded-xl border border-green-100">
                            <div className="text-green-800 text-xs font-bold uppercase tracking-wider mb-1">Validé</div>
                            <div className="text-xl font-serif font-bold text-green-900">
                                {new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'MAD', maximumFractionDigits: 0 }).format(totalValidated)}
                            </div>
                        </div>
                    </div>
                </div>

                {/* Transactions List */}
                <div className="md:col-span-2 space-y-6">
                    <h2 className="font-serif text-lg font-bold flex items-center gap-2">
                        <FileText className="w-5 h-5" /> Transactions ({lines.length})
                    </h2>

                    <div className="space-y-3">
                        {lines.map(line => (
                            <Link key={line.id} href={`/pending-lines/${line.id}`}>
                                <div className="bg-white p-4 rounded-xl border shadow-sm hover:shadow-md transition-shadow flex items-center justify-between group">
                                    <div className="flex items-center gap-4">
                                        <div className="w-10 h-10 rounded-lg bg-slate-50 border flex items-center justify-center text-slate-400 group-hover:bg-[#1A1A1A] group-hover:text-white transition-colors">
                                            <CreditCard className="w-5 h-5" />
                                        </div>
                                        <div>
                                            <div className="font-medium text-sm text-[#1A1A1A]">{line.bank_label}</div>
                                            <div className="text-xs text-muted-foreground flex items-center gap-2 mt-0.5">
                                                <Calendar className="w-3 h-3" /> {formatDate(line.transaction_date)}
                                            </div>
                                        </div>
                                    </div>
                                    <div className="text-right">
                                        <div className="font-bold tabular-nums text-sm">{formatAmount(line.amount)}</div>
                                        <div className="mt-1"><StatusBadge status={line.status} /></div>
                                    </div>
                                </div>
                            </Link>
                        ))}
                        {lines.length === 0 && (
                            <div className="text-center py-12 text-muted-foreground bg-white rounded-xl border border-dashed">
                                Aucune transaction associée.
                            </div>
                        )}
                    </div>
                </div>
            </main>
        </div>
    );
}
