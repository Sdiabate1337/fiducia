'use client';

import { useEffect, useState } from 'react';

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

export default function DashboardPage() {
    const [lines, setLines] = useState<PendingLine[]>([]);
    const [stats, setStats] = useState<Stats | null>(null);
    const [loading, setLoading] = useState(true);
    const [statusFilter, setStatusFilter] = useState<string>('all');
    const [search, setSearch] = useState('');

    // Demo cabinet ID
    const cabinetId = '00000000-0000-0000-0000-000000000001';

    useEffect(() => {
        fetchData();
    }, [statusFilter, search]);

    const fetchData = async () => {
        setLoading(true);
        try {
            // Fetch stats
            const statsRes = await fetch(`/api/v1/cabinets/${cabinetId}/pending-lines/stats`);
            if (statsRes.ok) {
                const statsData = await statsRes.json();
                setStats(statsData);
            }

            // Fetch lines with filters
            let url = `/api/v1/cabinets/${cabinetId}/pending-lines?limit=50`;
            if (statusFilter !== 'all') {
                url += `&status=${statusFilter}`;
            }
            if (search) {
                url += `&search=${encodeURIComponent(search)}`;
            }

            const linesRes = await fetch(url);
            if (linesRes.ok) {
                const linesData: ListResponse = await linesRes.json();
                setLines(linesData.items || []);
            }
        } catch (err) {
            console.error('Failed to fetch data:', err);
        } finally {
            setLoading(false);
        }
    };

    const formatAmount = (amount: string) => {
        const num = parseFloat(amount);
        return new Intl.NumberFormat('fr-FR', {
            style: 'currency',
            currency: 'EUR',
        }).format(num);
    };

    const formatDate = (dateStr: string) => {
        const date = new Date(dateStr);
        return date.toLocaleDateString('fr-FR', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
        });
    };

    const getStatusBadge = (status: string) => {
        const styles: Record<string, { class: string; label: string }> = {
            pending: { class: 'badge-pending', label: 'En attente' },
            contacted: { class: 'badge-pending', label: 'Contacté' },
            received: { class: 'badge-success', label: 'Reçu' },
            validated: { class: 'badge-success', label: 'Validé' },
            rejected: { class: 'badge-error', label: 'Rejeté' },
            expired: { class: 'badge-error', label: 'Expiré' },
        };
        const style = styles[status] || { class: '', label: status };
        return <span className={`badge ${style.class}`}>{style.label}</span>;
    };

    return (
        <main style={{ minHeight: '100vh', padding: '2rem' }}>
            {/* Header */}
            <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '2rem'
            }}>
                <div>
                    <a href="/" style={{ color: '#888', fontSize: '0.875rem' }}>← Accueil</a>
                    <h1 style={{ fontSize: '1.75rem', marginTop: '0.5rem' }}>
                        Dashboard
                    </h1>
                </div>
                <div style={{ display: 'flex', gap: '1rem' }}>
                    <a href="/settings" className="btn btn-secondary">
                        ⚙️ Paramètres
                    </a>
                    <a href="/import" className="btn btn-primary">
                        📤 Importer CSV
                    </a>
                </div>
            </div>

            {/* Stats Cards */}
            {stats && (
                <div style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))',
                    gap: '1rem',
                    marginBottom: '2rem'
                }}>
                    <div className="card" style={{ textAlign: 'center' }}>
                        <div style={{ fontSize: '1.75rem', fontWeight: 700 }}>{stats.total}</div>
                        <div style={{ color: '#888', fontSize: '0.75rem' }}>Total lignes</div>
                    </div>
                    <div
                        className="card"
                        style={{
                            textAlign: 'center',
                            cursor: 'pointer',
                            borderColor: statusFilter === 'pending' ? '#6366f1' : undefined
                        }}
                        onClick={() => setStatusFilter(statusFilter === 'pending' ? 'all' : 'pending')}
                    >
                        <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#f59e0b' }}>
                            {stats.pending}
                        </div>
                        <div style={{ color: '#888', fontSize: '0.75rem' }}>En attente</div>
                    </div>
                    <div
                        className="card"
                        style={{
                            textAlign: 'center',
                            cursor: 'pointer',
                            borderColor: statusFilter === 'contacted' ? '#6366f1' : undefined
                        }}
                        onClick={() => setStatusFilter(statusFilter === 'contacted' ? 'all' : 'contacted')}
                    >
                        <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#6366f1' }}>
                            {stats.contacted}
                        </div>
                        <div style={{ color: '#888', fontSize: '0.75rem' }}>Contactés</div>
                    </div>
                    <div
                        className="card"
                        style={{
                            textAlign: 'center',
                            cursor: 'pointer',
                            borderColor: statusFilter === 'received' ? '#6366f1' : undefined
                        }}
                        onClick={() => setStatusFilter(statusFilter === 'received' ? 'all' : 'received')}
                    >
                        <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#22c55e' }}>
                            {stats.received}
                        </div>
                        <div style={{ color: '#888', fontSize: '0.75rem' }}>Reçus</div>
                    </div>
                    <div
                        className="card"
                        style={{
                            textAlign: 'center',
                            cursor: 'pointer',
                            borderColor: statusFilter === 'validated' ? '#6366f1' : undefined
                        }}
                        onClick={() => setStatusFilter(statusFilter === 'validated' ? 'all' : 'validated')}
                    >
                        <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#22c55e' }}>
                            {stats.validated}
                        </div>
                        <div style={{ color: '#888', fontSize: '0.75rem' }}>Validés</div>
                    </div>
                </div>
            )}

            {/* Amount Summary */}
            {stats && (
                <div style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(2, 1fr)',
                    gap: '1rem',
                    marginBottom: '2rem'
                }}>
                    <div className="card">
                        <div style={{ color: '#888', fontSize: '0.75rem', marginBottom: '0.25rem' }}>
                            Montant en attente
                        </div>
                        <div style={{ fontSize: '1.5rem', fontWeight: 600, color: '#f59e0b' }}>
                            {formatAmount(stats.pending_amount.toString())}
                        </div>
                    </div>
                    <div className="card">
                        <div style={{ color: '#888', fontSize: '0.75rem', marginBottom: '0.25rem' }}>
                            Montant validé
                        </div>
                        <div style={{ fontSize: '1.5rem', fontWeight: 600, color: '#22c55e' }}>
                            {formatAmount(stats.validated_amount.toString())}
                        </div>
                    </div>
                </div>
            )}

            {/* Filters */}
            <div style={{
                display: 'flex',
                gap: '1rem',
                marginBottom: '1rem',
                flexWrap: 'wrap'
            }}>
                <input
                    className="input"
                    style={{ maxWidth: '300px' }}
                    placeholder="Rechercher par libellé..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                />
                <select
                    className="input"
                    style={{ maxWidth: '200px' }}
                    value={statusFilter}
                    onChange={(e) => setStatusFilter(e.target.value)}
                >
                    <option value="all">Tous les statuts</option>
                    <option value="pending">En attente</option>
                    <option value="contacted">Contacté</option>
                    <option value="received">Reçu</option>
                    <option value="validated">Validé</option>
                    <option value="rejected">Rejeté</option>
                </select>
                {statusFilter !== 'all' && (
                    <button
                        className="btn btn-secondary"
                        onClick={() => setStatusFilter('all')}
                    >
                        Effacer filtres
                    </button>
                )}
            </div>

            {/* Lines Table */}
            <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
                {loading ? (
                    <div style={{ padding: '3rem', textAlign: 'center' }} className="animate-pulse">
                        Chargement...
                    </div>
                ) : lines.length === 0 ? (
                    <div style={{ padding: '3rem', textAlign: 'center', color: '#888' }}>
                        <div style={{ fontSize: '2rem', marginBottom: '1rem' }}>📭</div>
                        <p>Aucune ligne trouvée</p>
                        <a href="/import" className="btn btn-primary" style={{ marginTop: '1rem' }}>
                            Importer des données
                        </a>
                    </div>
                ) : (
                    <div style={{ overflowX: 'auto' }}>
                        <table>
                            <thead>
                                <tr>
                                    <th>Date</th>
                                    <th>Libellé</th>
                                    <th>Client</th>
                                    <th style={{ textAlign: 'right' }}>Montant</th>
                                    <th>Statut</th>
                                    <th></th>
                                </tr>
                            </thead>
                            <tbody>
                                {lines.map((line) => (
                                    <tr key={line.id}>
                                        <td>{formatDate(line.transaction_date)}</td>
                                        <td style={{ maxWidth: '300px' }}>
                                            <div style={{
                                                whiteSpace: 'nowrap',
                                                overflow: 'hidden',
                                                textOverflow: 'ellipsis'
                                            }}>
                                                {line.bank_label || '—'}
                                            </div>
                                        </td>
                                        <td>
                                            {line.client ? (
                                                <span>{line.client.name}</span>
                                            ) : (
                                                <span style={{ color: '#666' }}>Non assigné</span>
                                            )}
                                        </td>
                                        <td style={{ textAlign: 'right', fontFamily: 'monospace' }}>
                                            {formatAmount(line.amount)}
                                        </td>
                                        <td>{getStatusBadge(line.status)}</td>
                                        <td>
                                            <a href={`/pending-lines/${line.id}`} className="btn btn-secondary" style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem', textDecoration: 'none' }}>
                                                Voir
                                            </a>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </main>
    );
}
