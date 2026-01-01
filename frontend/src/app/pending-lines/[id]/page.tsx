'use client';

import { useEffect, useState, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { motion, AnimatePresence } from 'framer-motion';
import {
    ArrowLeft, Send, Mic, MessageSquare, FileText, CheckCircle2,
    AlertTriangle, Calendar, Euro, Building2, Hash, Clock,
    ChevronRight, Paperclip, Download, Maximize2, Shield, Eye,
    Search, Settings, MessageCircle, Layout, Layers
} from 'lucide-react';
import Link from 'next/link';
import { cn } from '@/lib/utils'; // Assuming cn utility exists or I will implement inline

// --- Types ---

interface Message {
    id: string;
    direction: string;
    message_type: string;
    content: string | null;
    status: string;
    created_at: string;
    sent_at: string | null;
    delivered_at: string | null;
    read_at: string | null;
}

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
}

interface Document {
    id: string;
    file_path: string;
    file_type: string | null;
    ocr_status: string;
    ocr_text: string | null;
    ocr_data: {
        date?: string;
        amount?: number;
        vendor?: string;
        invoice_number?: string;
        document_type?: string;
        confidence?: number;
    } | null;
    match_confidence: string;
    match_status: string;
    created_at: string;
}

// --- Components ---

const StatusBadge = ({ status }: { status: string }) => {
    const styles: any = {
        pending: { color: 'bg-amber-100 text-amber-800 border-amber-200', label: 'En attente', icon: Clock },
        contacted: { color: 'bg-blue-100 text-blue-800 border-blue-200', label: 'Contacté', icon: Send },
        received: { color: 'bg-purple-100 text-purple-800 border-purple-200', label: 'Reçu', icon: FileText },
        validated: { color: 'bg-green-100 text-green-800 border-green-200', label: 'Validé', icon: CheckCircle2 },
        rejected: { color: 'bg-red-100 text-red-800 border-red-200', label: 'Rejeté', icon: AlertTriangle },
    };
    const style = styles[status] || styles.pending;
    const Icon = style.icon;

    return (
        <span className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-[10px] md:text-xs font-bold uppercase tracking-wider border ${style.color}`}>
            <Icon size={12} strokeWidth={2.5} />
            {style.label}
        </span>
    );
};

export default function PendingLineDetailPage() {
    const params = useParams();
    const router = useRouter();
    const id = params.id as string;
    const scrollRef = useRef<HTMLDivElement>(null);

    const [line, setLine] = useState<PendingLine | null>(null);
    const [messages, setMessages] = useState<Message[]>([]);
    const [documents, setDocuments] = useState<Document[]>([]);
    const [loading, setLoading] = useState(true);
    const [sending, setSending] = useState(false);
    const [customMessage, setCustomMessage] = useState('');
    const [messageType, setMessageType] = useState<'text' | 'voice'>('text');
    const [activeTab, setActiveTab] = useState<'history' | 'action'>('action');

    // Mobile View State: 'doc' or 'console'
    const [mobileView, setMobileView] = useState<'doc' | 'console'>('console');

    useEffect(() => {
        fetchData();
    }, [id]);

    // Scroll to bottom of chat when messages change
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [messages, activeTab, mobileView]);

    const fetchData = async () => {
        setLoading(true);
        try {
            const [lineRes, msgRes, docRes] = await Promise.all([
                fetch(`/api/v1/pending-lines/${id}`),
                fetch(`/api/v1/pending-lines/${id}/messages`),
                fetch(`/api/v1/pending-lines/${id}/documents`)
            ]);

            if (lineRes.ok) setLine(await lineRes.json());
            if (msgRes.ok) setMessages((await msgRes.json()).messages || []);
            if (docRes.ok) {
                const docs = (await docRes.json()).documents || [];
                setDocuments(docs);
                // If documents exist, default mobile view to doc? Or stick to console?
                // Let's stick to console as action is primary, but user can switch.
            }
        } catch (err) {
            console.error('Failed to fetch data:', err);
        } finally {
            setLoading(false);
        }
    };

    const approveDocument = async (docId: string) => {
        try {
            const res = await fetch(`/api/v1/documents/${docId}/approve`, { method: 'POST' });
            if (res.ok) {
                if (window.navigator && window.navigator.vibrate) {
                    window.navigator.vibrate(200);
                }
                fetchData();
            } else {
                const err = await res.json();
                alert('Erreur: ' + (err.error || 'Échec de la validation'));
            }
        } catch (err) {
            alert('Erreur réseau');
        }
    };

    const sendRelance = async (immediate: boolean = false) => {
        setSending(true);
        try {
            const res = await fetch(`/api/v1/pending-lines/${id}/messages`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    message_type: messageType,
                    custom_message: customMessage || undefined,
                    immediate,
                }),
            });

            if (res.ok) {
                setCustomMessage('');
                fetchData();
                setActiveTab('history'); // Switch to history to see the new message
            } else {
                alert('Échec de l\'envoi');
            }
        } catch (err) {
            alert('Erreur réseau');
        } finally {
            setSending(false);
        }
    };

    const formatDate = (dateStr: string) => {
        return new Date(dateStr).toLocaleString('fr-FR', {
            day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit'
        });
    };

    const formatAmount = (amount: string) => {
        const num = parseFloat(amount);
        return new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'MAD' }).format(num);
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-[#F9F8F6] flex items-center justify-center">
                <div className="text-center">
                    <div className="animate-spin w-8 h-8 border-2 border-[#1A1A1A] border-t-transparent rounded-full mx-auto mb-4" />
                    <p className="text-[#1A1A1A]/40 font-medium animate-pulse">Chargement du dossier...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="h-screen bg-[#F9F8F6] text-[#1A1A1A] font-sans selection:bg-[#4F2830]/20 flex flex-col overflow-hidden">
            {/* Navbar */}
            <nav className="shrink-0 bg-[#F9F8F6] border-b border-[#1A1A1A]/5 px-4 md:px-8 h-16 flex items-center justify-between z-50">
                <div className="flex items-center gap-4 md:gap-6">
                    <Link href="/dashboard" className="p-2 -ml-2 text-[#1A1A1A]/40 hover:text-[#1A1A1A] transition-colors rounded-lg hover:bg-[#1A1A1A]/5">
                        <ArrowLeft size={20} />
                    </Link>
                    <div className="flex items-center gap-2 md:gap-4 truncate">
                        <span className="text-lg md:text-xl font-serif font-bold tracking-tight truncate">
                            Dossier #{id.substring(0, 4)}...
                        </span>
                        {line && <div className="hidden md:block"><StatusBadge status={line.status} /></div>}
                    </div>
                </div>
                <div className="flex items-center gap-3">
                    <div className="text-right mr-2 hidden md:block">
                        <div className="text-xs text-[#1A1A1A]/40 font-medium uppercase tracking-wider">Client</div>
                        <div className="font-serif font-bold">{line?.client?.name || 'Non assigné'}</div>
                    </div>
                    {line && <div className="md:hidden"><StatusBadge status={line.status} /></div>}
                    <div className="w-8 h-8 rounded-full bg-[#1A1A1A] text-white flex items-center justify-center font-serif text-sm">
                        {line?.client?.name?.substring(0, 2).toUpperCase() || 'NA'}
                    </div>
                </div>
            </nav>

            <div className="flex-1 flex overflow-hidden relative">

                {/* Left Panel: Document Viewer (Lens) - Hidden on Mobile unless selected */}
                <div className={`
                    w-full md:w-1/2 md:border-r md:border-[#1A1A1A]/5 bg-[#EAE8E4]/30 p-4 md:p-8 flex flex-col overflow-y-auto
                    ${mobileView === 'doc' ? 'flex' : 'hidden md:flex'}
                `}>
                    <div className="mb-4 md:mb-6 flex justify-between items-center">
                        <h2 className="text-sm font-bold uppercase tracking-widest text-[#1A1A1A]/40">Pièce Comptable</h2>
                        <button className="text-xs font-medium text-[#1A1A1A]/60 hover:text-[#1A1A1A] flex items-center gap-1 transition-colors">
                            <Maximize2 size={12} /> <span className="hidden md:inline">Plein écran</span>
                        </button>
                    </div>

                    <div className="flex-1 flex flex-col gap-6 pb-20 md:pb-0">
                        {documents.length > 0 ? (
                            documents.map((doc) => (
                                <motion.div
                                    key={doc.id}
                                    initial={{ opacity: 0, y: 20 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    className="relative group bg-white rounded-xl shadow-2xl shadow-[#1A1A1A]/10 overflow-hidden border border-[#1A1A1A]/5"
                                >
                                    {/* Document Header Overlay */}
                                    <div className="absolute top-0 inset-x-0 p-4 bg-gradient-to-b from-black/50 to-transparent z-10 flex justify-between items-start">
                                        <span className="bg-white/90 backdrop-blur text-[#1A1A1A] text-[10px] font-bold px-2 py-1 rounded uppercase tracking-wider">
                                            {doc.ocr_data?.document_type === 'invoice' ? 'Facture' : 'Reçu'}
                                        </span>
                                    </div>

                                    {/* Simulated Preview Area */}
                                    <div className="aspect-[3/4] bg-[#333] relative flex items-center justify-center overflow-hidden">
                                        <img
                                            src={`/api/v1/documents/content/${doc.file_path.split('/').pop()}`}
                                            alt="Justificatif"
                                            className="w-full h-full object-contain relative z-10"
                                            onError={(e) => {
                                                e.currentTarget.style.display = 'none';
                                                e.currentTarget.nextElementSibling?.classList.remove('hidden');
                                            }}
                                        />
                                        {/* Fallback Placeholder */}
                                        <div className="text-center text-white/20 absolute inset-0 flex flex-col items-center justify-center hidden">
                                            <FileText className="w-12 h-12 md:w-16 md:h-16 mx-auto mb-4 opacity-50" />
                                            <p className="font-serif text-base md:text-lg">Aperçu du document</p>
                                        </div>
                                    </div>

                                    {/* AI Confidence / Action Footer */}
                                    <div className="bg-white p-4 md:p-6 border-t border-[#1A1A1A]/5">
                                        <div className="flex justify-between items-end">
                                            <div>
                                                <div className="text-[#1A1A1A]/40 text-xs uppercase tracking-wider font-bold mb-1">Confiance IA</div>
                                                <div className={`text-2xl md:text-3xl font-serif ${parseFloat(doc.match_confidence) >= 0.8 ? 'text-[#1A4D2E]' : 'text-amber-500'}`}>
                                                    {(parseFloat(doc.match_confidence) * 100).toFixed(0)}%
                                                </div>
                                            </div>

                                            {doc.match_status === 'pending' && (
                                                <button
                                                    onClick={() => approveDocument(doc.id)}
                                                    className="px-4 md:px-6 py-2 md:py-3 bg-[#1A1A1A] text-white rounded-lg font-medium hover:bg-[#1A4D2E] transition-all flex items-center gap-2 shadow-lg hover:shadow-xl hover:-translate-y-0.5 text-sm md:text-base"
                                                >
                                                    <CheckCircle2 size={16} />
                                                    <span className="hidden md:inline">Valider le match</span>
                                                    <span className="md:hidden">Valider</span>
                                                </button>
                                            )}
                                            {doc.match_status === 'auto_matched' && (
                                                <div className="flex items-center gap-2 text-[#1A4D2E] font-medium bg-[#1A4D2E]/5 px-3 py-2 md:px-4 rounded-lg text-sm md:text-base">
                                                    <Shield size={16} />
                                                    <span className="hidden md:inline">Auto-validé par IA</span>
                                                    <span className="md:hidden">Auto-validé</span>
                                                </div>
                                            )}
                                        </div>

                                        {doc.ocr_data && (
                                            <div className="mt-4 md:mt-6 pt-4 md:pt-6 border-t border-[#1A1A1A]/5 grid grid-cols-2 gap-4">
                                                <div>
                                                    <div className="text-[10px] uppercase text-[#1A1A1A]/40 font-bold">Vendeur détecté</div>
                                                    <div className="font-medium text-[#1A1A1A] text-sm md:text-base">{doc.ocr_data.vendor || 'Inconnu'}</div>
                                                </div>
                                                <div>
                                                    <div className="text-[10px] uppercase text-[#1A1A1A]/40 font-bold">Montant détecté</div>
                                                    <div className="font-medium text-[#1A1A1A] text-sm md:text-base">{doc.ocr_data.amount ? formatAmount(doc.ocr_data.amount.toString()) : '—'}</div>
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                </motion.div>
                            ))
                        ) : (
                            <div className="h-[300px] md:h-[400px] border-2 border-dashed border-[#1A1A1A]/10 rounded-2xl flex flex-col items-center justify-center text-center p-8">
                                <div className="w-12 h-12 md:w-16 md:h-16 bg-[#1A1A1A]/5 rounded-full flex items-center justify-center mb-4 text-[#1A1A1A]/30">
                                    <Paperclip className="w-5 h-5 md:w-6 md:h-6" />
                                </div>
                                <h3 className="text-base md:text-lg font-serif font-medium text-[#1A1A1A] mb-2">Aucun justificatif</h3>
                                <p className="text-sm text-[#1A1A1A]/50 max-w-xs leading-relaxed">
                                    Le client n'a pas encore envoyé de photo pour cette transaction.
                                </p>
                            </div>
                        )}
                    </div>
                </div>

                {/* Right Panel: Context & Action Console - Hidden on Mobile unless selected */}
                <div className={`
                    w-full md:w-1/2 bg-white flex flex-col
                    ${mobileView === 'console' ? 'flex' : 'hidden md:flex'}
                `}>
                    {/* Transaction Ticket */}
                    <div className="p-6 md:p-8 border-b border-[#1A1A1A]/5">
                        <div className="flex justify-between items-start mb-6">
                            <div>
                                <h2 className="text-2xl md:text-3xl font-serif text-[#1A1A1A] mb-1">{line ? formatAmount(line.amount) : '—'}</h2>
                                <p className="text-[#1A1A1A]/50 font-medium text-sm md:text-base">{line?.bank_label}</p>
                            </div>
                            <div className="text-right">
                                <div className="text-sm font-bold text-[#1A1A1A]">{line?.transaction_date ? new Date(line.transaction_date).toLocaleDateString() : '—'}</div>
                                <div className="text-xs text-[#1A1A1A]/40 uppercase tracking-wide mt-1">Date Transaction</div>
                            </div>
                        </div>

                        {/* Tabs */}
                        <div className="flex gap-6 md:gap-8 border-b border-[#1A1A1A]/5 overflow-x-auto">
                            <button
                                onClick={() => setActiveTab('action')}
                                className={`pb-4 text-sm font-bold uppercase tracking-wider transition-all relative whitespace-nowrap ${activeTab === 'action' ? 'text-[#1A1A1A]' : 'text-[#1A1A1A]/40 hover:text-[#1A1A1A]'}`}
                            >
                                Action requise
                                {activeTab === 'action' && <motion.div layoutId="tab" className="absolute bottom-0 left-0 right-0 h-0.5 bg-[#1A1A1A]" />}
                            </button>
                            <button
                                onClick={() => setActiveTab('history')}
                                className={`pb-4 text-sm font-bold uppercase tracking-wider transition-all relative whitespace-nowrap ${activeTab === 'history' ? 'text-[#1A1A1A]' : 'text-[#1A1A1A]/40 hover:text-[#1A1A1A]'}`}
                            >
                                Historique
                                {messages.length > 0 && <span className="ml-2 px-1.5 py-0.5 bg-[#1A1A1A]/5 rounded-full text-[10px]">{messages.length}</span>}
                                {activeTab === 'history' && <motion.div layoutId="tab" className="absolute bottom-0 left-0 right-0 h-0.5 bg-[#1A1A1A]" />}
                            </button>
                        </div>
                    </div>

                    {/* Tab Content */}
                    <div className="flex-1 overflow-hidden relative bg-[#F9F8F6]/30 pb-20 md:pb-0">
                        <AnimatePresence mode="wait">
                            {activeTab === 'action' ? (
                                <motion.div
                                    key="action"
                                    initial={{ opacity: 0, x: 20 }}
                                    animate={{ opacity: 1, x: 0 }}
                                    exit={{ opacity: 0, x: -20 }}
                                    className="h-full p-4 md:p-8 flex flex-col"
                                >
                                    <div className="bg-white border boundary-[#1A1A1A]/10 rounded-2xl p-1 shadow-sm mb-4 md:mb-6 flex gap-1 w-full md:w-fit">
                                        <button
                                            onClick={() => setMessageType('text')}
                                            className={`flex-1 md:flex-none px-4 py-2 rounded-xl text-xs md:text-sm font-medium transition-all flex items-center justify-center gap-2 ${messageType === 'text' ? 'bg-[#1A1A1A] text-white shadow-md' : 'text-[#1A1A1A]/60 hover:text-[#1A1A1A]'}`}
                                        >
                                            <MessageCircle size={16} /> WhatsApp
                                        </button>
                                        <button
                                            onClick={() => setMessageType('voice')}
                                            className={`flex-1 md:flex-none px-4 py-2 rounded-xl text-xs md:text-sm font-medium transition-all flex items-center justify-center gap-2 ${messageType === 'voice' ? 'bg-[#1A1A1A] text-white shadow-md' : 'text-[#1A1A1A]/60 hover:text-[#1A1A1A]'}`}
                                        >
                                            <Mic size={16} /> Vocal IA
                                        </button>
                                    </div>

                                    <div className="flex-1 mb-4 md:mb-6">
                                        {messageType === 'text' ? (
                                            <textarea
                                                className="w-full h-full p-4 md:p-6 bg-white border border-[#1A1A1A]/10 rounded-2xl focus:border-[#1A1A1A]/30 focus:ring-4 focus:ring-[#1A1A1A]/5 outline-none text-base resize-none transition-all placeholder:text-[#1A1A1A]/20"
                                                placeholder="Rédigez votre demande de justificatif ici..."
                                                value={customMessage}
                                                onChange={(e) => setCustomMessage(e.target.value)}
                                            />
                                        ) : (
                                            <div className="w-full h-full bg-[#1A1A1A] rounded-2xl flex flex-col items-center justify-center text-white relative overflow-hidden">
                                                <div className="absolute inset-0 bg-[#4F2830]/20 mix-blend-overlay" />
                                                <div className="relative z-10 flex flex-col items-center">
                                                    <div className="flex items-end gap-1 mb-4 h-12">
                                                        {[...Array(8)].map((_, i) => (
                                                            <motion.div
                                                                key={i}
                                                                animate={{ height: [10, 32, 10] }}
                                                                transition={{ repeat: Infinity, duration: 1.5, delay: i * 0.1, ease: "easeInOut" }}
                                                                className="w-1.5 md:w-2 bg-white/80 rounded-full"
                                                            />
                                                        ))}
                                                    </div>
                                                    <p className="font-serif text-lg md:text-xl mb-1">Génération de message vocal</p>
                                                    <p className="text-white/40 text-xs md:text-sm">Utilise votre voix clonée</p>
                                                </div>
                                            </div>
                                        )}
                                    </div>

                                    <div className="flex flex-col md:flex-row gap-3 md:gap-4">
                                        <button
                                            onClick={() => sendRelance(false)}
                                            disabled={sending}
                                            className="w-full md:flex-1 py-3 md:py-4 bg-white border border-[#1A1A1A]/10 text-[#1A1A1A] font-medium rounded-xl hover:bg-[#F9F8F6] transition-all flex items-center justify-center gap-2 text-sm md:text-base"
                                        >
                                            <Clock size={18} />
                                            Programmer (Anti-ban)
                                        </button>
                                        <button
                                            onClick={() => sendRelance(true)}
                                            disabled={sending}
                                            className="w-full md:flex-[2] py-3 md:py-4 bg-[#1A1A1A] text-white font-medium rounded-xl hover:bg-[#1A4D2E] transition-all flex items-center justify-center gap-2 shadow-xl hover:shadow-2xl hover:-translate-y-1 transform disabled:opacity-70 text-sm md:text-base"
                                        >
                                            <Send size={18} />
                                            {sending ? 'Envoi...' : 'Envoyer Immédiatement'}
                                        </button>
                                    </div>
                                </motion.div>
                            ) : (
                                <motion.div
                                    key="history"
                                    initial={{ opacity: 0, x: -20 }}
                                    animate={{ opacity: 1, x: 0 }}
                                    exit={{ opacity: 0, x: 20 }}
                                    className="h-full overflow-y-auto p-4 md:p-8"
                                    ref={scrollRef}
                                >
                                    <div className="space-y-6 md:space-y-8">
                                        {messages.length === 0 ? (
                                            <div className="text-center py-20 opacity-40">
                                                <MessageSquare size={40} className="mx-auto mb-4" />
                                                <p>Aucun message échangé pour le moment.</p>
                                            </div>
                                        ) : (
                                            messages.map((msg) => (
                                                <div key={msg.id} className={`flex ${msg.direction === 'outbound' ? 'justify-end' : 'justify-start'}`}>
                                                    <div className={`max-w-[85%] md:max-w-[80%] ${msg.direction === 'outbound' ? 'items-end' : 'items-start'} flex flex-col`}>
                                                        <div className={`p-4 rounded-2xl text-sm leading-relaxed ${msg.direction === 'outbound'
                                                            ? 'bg-[#1A1A1A] text-white rounded-br-none'
                                                            : 'bg-white border border-[#1A1A1A]/10 text-[#1A1A1A] rounded-bl-none shadow-sm'
                                                            }`}>
                                                            {msg.message_type === 'voice' ? (
                                                                <div className="flex items-center gap-3">
                                                                    <div className="w-8 h-8 rounded-full bg-white/20 flex items-center justify-center">
                                                                        <Mic size={14} />
                                                                    </div>
                                                                    <span>Message Vocal (0:34)</span>
                                                                </div>
                                                            ) : (
                                                                msg.content || <span className="italic opacity-70">Document partagé</span>
                                                            )}
                                                        </div>
                                                        <div className="mt-2 text-[10px] text-[#1A1A1A]/40 font-medium uppercase tracking-wider flex items-center gap-2">
                                                            {msg.direction === 'outbound' ? 'Vous' : 'Client'} • {formatDate(msg.created_at)}
                                                        </div>
                                                    </div>
                                                </div>
                                            ))
                                        )}
                                    </div>
                                </motion.div>
                            )}
                        </AnimatePresence>
                    </div>
                </div>

                {/* Mobile Bottom Navigation - Tabs for View Switching */}
                <div className="md:hidden absolute bottom-0 inset-x-0 bg-white border-t border-[#1A1A1A]/10 p-2 flex justify-around z-50 shadow-[0_-5px_15px_rgba(0,0,0,0.05)]">
                    <button
                        onClick={() => setMobileView('console')}
                        className={`flex flex-col items-center justify-center p-2 rounded-xl flex-1 transition-colors ${mobileView === 'console' ? 'text-[#1A1A1A] bg-[#1A1A1A]/5' : 'text-[#1A1A1A]/40'}`}
                    >
                        <Layers size={20} className="mb-1" />
                        <span className="text-[10px] font-bold uppercase tracking-wider">Console</span>
                    </button>
                    <button
                        onClick={() => setMobileView('doc')}
                        className={`flex flex-col items-center justify-center p-2 rounded-xl flex-1 transition-colors ${mobileView === 'doc' ? 'text-[#1A1A1A] bg-[#1A1A1A]/5' : 'text-[#1A1A1A]/40'}`}
                    >
                        <div className="relative">
                            <FileText size={20} className="mb-1" />
                            {documents.length > 0 && <span className="absolute -top-1 -right-1 w-2.5 h-2.5 bg-[#20E070] rounded-full border-2 border-white" />}
                        </div>
                        <span className="text-[10px] font-bold uppercase tracking-wider">Documents</span>
                    </button>
                </div>

            </div>
        </div>
    );
}
