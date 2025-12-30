'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';

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

export default function PendingLineDetailPage() {
    const params = useParams();
    const router = useRouter();
    const id = params.id as string;

    const [line, setLine] = useState<PendingLine | null>(null);
    const [messages, setMessages] = useState<Message[]>([]);
    const [loading, setLoading] = useState(true);
    const [sending, setSending] = useState(false);
    const [customMessage, setCustomMessage] = useState('');

    useEffect(() => {
        fetchData();
    }, [id]);

    const fetchData = async () => {
        setLoading(true);
        try {
            // Fetch pending line
            const lineRes = await fetch(`/api/v1/pending-lines/${id}`);
            if (lineRes.ok) {
                const lineData = await lineRes.json();
                setLine(lineData);
            }

            // Fetch messages
            const msgRes = await fetch(`/api/v1/pending-lines/${id}/messages`);
            if (msgRes.ok) {
                const msgData = await msgRes.json();
                setMessages(msgData.messages || []);
            }
        } catch (err) {
            console.error('Failed to fetch data:', err);
        } finally {
            setLoading(false);
        }
    };

    const sendRelance = async (immediate: boolean = false) => {
        setSending(true);
        try {
            const res = await fetch(`/api/v1/pending-lines/${id}/messages`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    message_type: 'text',
                    custom_message: customMessage || undefined,
                    immediate,
                }),
            });

            if (res.ok) {
                const data = await res.json();
                alert(`Relance ${immediate ? 'envoyée' : 'programmée'} ! ID: ${data.id}`);
                setCustomMessage('');
                fetchData(); // Refresh messages
            } else {
                const err = await res.json();
                alert('Erreur: ' + (err.error || 'Échec de l\'envoi'));
            }
        } catch (err) {
            alert('Erreur réseau');
        } finally {
            setSending(false);
        }
    };

    const formatDate = (dateStr: string) => {
        return new Date(dateStr).toLocaleString('fr-FR', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    const formatAmount = (amount: string) => {
        const num = parseFloat(amount);
        return new Intl.NumberFormat('fr-FR', {
            style: 'currency',
            currency: 'EUR',
        }).format(num);
    };

    const getStatusBadge = (status: string) => {
        const styles: Record<string, { class: string; label: string }> = {
            queued: { class: 'badge-pending', label: '⏳ En attente' },
            sending: { class: 'badge-pending', label: '📤 Envoi...' },
            sent: { class: 'badge-success', label: '✓ Envoyé' },
            delivered: { class: 'badge-success', label: '✓✓ Délivré' },
            read: { class: 'badge-success', label: '👁 Lu' },
            failed: { class: 'badge-error', label: '✗ Échec' },
        };
        const style = styles[status] || { class: '', label: status };
        return <span className={`badge ${style.class}`}>{style.label}</span>;
    };

    if (loading) {
        return (
            <main style={{ minHeight: '100vh', padding: '2rem' }}>
                <div className="animate-pulse">Chargement...</div>
            </main>
        );
    }

    return (
        <main style={{ minHeight: '100vh', padding: '2rem' }}>
            {/* Header */}
            <div style={{ marginBottom: '1.5rem' }}>
                <button
                    onClick={() => router.back()}
                    style={{ color: '#888', fontSize: '0.875rem', background: 'none', border: 'none', cursor: 'pointer' }}
                >
                    ← Retour
                </button>
                <h1 style={{ fontSize: '1.5rem', marginTop: '0.5rem' }}>
                    Détail Ligne 471
                </h1>
            </div>

            {/* Line Info Card */}
            {line && (
                <div className="card" style={{ marginBottom: '1.5rem' }}>
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: '1rem' }}>
                        <div>
                            <div style={{ color: '#888', fontSize: '0.75rem' }}>Date</div>
                            <div>{formatDate(line.transaction_date).split(' ')[0]}</div>
                        </div>
                        <div>
                            <div style={{ color: '#888', fontSize: '0.75rem' }}>Libellé</div>
                            <div>{line.bank_label || '—'}</div>
                        </div>
                        <div>
                            <div style={{ color: '#888', fontSize: '0.75rem' }}>Montant</div>
                            <div style={{ fontWeight: 600, fontSize: '1.25rem' }}>{formatAmount(line.amount)}</div>
                        </div>
                        <div>
                            <div style={{ color: '#888', fontSize: '0.75rem' }}>Client</div>
                            <div>{line.client?.name || 'Non assigné'}</div>
                            {line.client?.phone && (
                                <div style={{ color: '#888', fontSize: '0.75rem' }}>{line.client.phone}</div>
                            )}
                        </div>
                    </div>
                </div>
            )}

            {/* Send Relance Section */}
            <div className="card" style={{ marginBottom: '1.5rem' }}>
                <h3 style={{ fontSize: '1rem', marginBottom: '1rem' }}>📤 Envoyer une relance</h3>

                {!line?.client ? (
                    <div style={{ color: '#888', padding: '1rem', textAlign: 'center' }}>
                        ⚠️ Assignez un client pour envoyer une relance
                    </div>
                ) : !line?.client?.phone ? (
                    <div style={{ color: '#888', padding: '1rem', textAlign: 'center' }}>
                        ⚠️ Le client n'a pas de numéro de téléphone
                    </div>
                ) : (
                    <>
                        <textarea
                            className="input"
                            placeholder="Message personnalisé (optionnel)..."
                            value={customMessage}
                            onChange={(e) => setCustomMessage(e.target.value)}
                            rows={3}
                            style={{ marginBottom: '1rem', resize: 'vertical' }}
                        />
                        <div style={{ display: 'flex', gap: '1rem' }}>
                            <button
                                className="btn btn-primary"
                                onClick={() => sendRelance(false)}
                                disabled={sending}
                                style={{ flex: 1 }}
                            >
                                {sending ? '⏳ Envoi...' : '📱 Programmer avec anti-ban'}
                            </button>
                            <button
                                className="btn btn-secondary"
                                onClick={() => sendRelance(true)}
                                disabled={sending}
                            >
                                ⚡ Envoyer immédiat
                            </button>
                        </div>
                        <p style={{ color: '#666', fontSize: '0.75rem', marginTop: '0.5rem' }}>
                            Le mode anti-ban ajoute un délai aléatoire de 30-180 secondes
                        </p>
                    </>
                )}
            </div>

            {/* Messages History */}
            <div className="card">
                <h3 style={{ fontSize: '1rem', marginBottom: '1rem' }}>
                    💬 Historique des messages ({messages.length})
                </h3>

                {messages.length === 0 ? (
                    <div style={{ color: '#888', textAlign: 'center', padding: '2rem' }}>
                        Aucun message envoyé
                    </div>
                ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                        {messages.map((msg) => (
                            <div
                                key={msg.id}
                                style={{
                                    padding: '0.75rem',
                                    borderRadius: '0.5rem',
                                    background: msg.direction === 'outbound' ? 'rgba(99, 102, 241, 0.1)' : 'rgba(34, 197, 94, 0.1)',
                                    marginLeft: msg.direction === 'outbound' ? '2rem' : 0,
                                    marginRight: msg.direction === 'inbound' ? '2rem' : 0,
                                }}
                            >
                                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.25rem' }}>
                                    <span style={{ color: '#888', fontSize: '0.75rem' }}>
                                        {msg.direction === 'outbound' ? '→ Envoyé' : '← Reçu'}
                                    </span>
                                    {getStatusBadge(msg.status)}
                                </div>
                                <div style={{ fontSize: '0.875rem' }}>
                                    {msg.content || '(message média)'}
                                </div>
                                <div style={{ color: '#666', fontSize: '0.7rem', marginTop: '0.25rem' }}>
                                    {formatDate(msg.created_at)}
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </main>
    );
}
