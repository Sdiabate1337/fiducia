'use client';

import { useEffect, useState } from 'react';

interface HealthStatus {
    status: string;
    service: string;
}

export default function Home() {
    const [health, setHealth] = useState<HealthStatus | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetch('/api/v1/health')
            .then(res => res.json())
            .then(data => {
                setHealth(data);
                setLoading(false);
            })
            .catch(() => setLoading(false));
    }, []);

    return (
        <main style={{
            minHeight: '100vh',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '2rem',
            gap: '2rem'
        }}>
            {/* Logo & Title */}
            <div style={{ textAlign: 'center' }}>
                <h1 style={{
                    fontSize: '3rem',
                    fontWeight: 700,
                    background: 'linear-gradient(135deg, #6366f1, #a855f7)',
                    WebkitBackgroundClip: 'text',
                    WebkitTextFillColor: 'transparent',
                    marginBottom: '0.5rem'
                }}>
                    Fiducia
                </h1>
                <p style={{ color: '#888', fontSize: '1.125rem' }}>
                    Assistant de Production Comptable Zero-Friction
                </p>
            </div>

            {/* Status Card */}
            <div className="card animate-fadeIn" style={{
                maxWidth: '400px',
                width: '100%',
                textAlign: 'center'
            }}>
                <h2 style={{ fontSize: '1.25rem', marginBottom: '1rem' }}>
                    État du Système
                </h2>

                {loading ? (
                    <div className="animate-pulse" style={{ color: '#888' }}>
                        Connexion au serveur...
                    </div>
                ) : health ? (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                        <div style={{
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            gap: '0.5rem'
                        }}>
                            <span style={{
                                width: '10px',
                                height: '10px',
                                borderRadius: '50%',
                                background: health.status === 'healthy' ? '#22c55e' : '#ef4444'
                            }} />
                            <span className={`badge ${health.status === 'healthy' ? 'badge-success' : 'badge-error'}`}>
                                {health.status === 'healthy' ? 'Opérationnel' : 'Erreur'}
                            </span>
                        </div>
                        <p style={{ color: '#888', fontSize: '0.875rem' }}>
                            Service: {health.service}
                        </p>
                    </div>
                ) : (
                    <div className="badge badge-error">
                        Serveur inaccessible
                    </div>
                )}
            </div>

            {/* Quick Links */}
            <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
                gap: '1rem',
                maxWidth: '800px',
                width: '100%'
            }}>
                <a href="/dashboard" className="card" style={{
                    textAlign: 'center',
                    transition: 'transform 0.2s, border-color 0.2s',
                    cursor: 'pointer'
                }}>
                    <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>📊</div>
                    <h3 style={{ fontSize: '1rem', marginBottom: '0.25rem' }}>Dashboard</h3>
                    <p style={{ color: '#888', fontSize: '0.75rem' }}>Vue d&apos;ensemble des lignes 471</p>
                </a>

                <a href="/import" className="card" style={{
                    textAlign: 'center',
                    transition: 'transform 0.2s, border-color 0.2s',
                    cursor: 'pointer'
                }}>
                    <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>📤</div>
                    <h3 style={{ fontSize: '1rem', marginBottom: '0.25rem' }}>Import CSV</h3>
                    <p style={{ color: '#888', fontSize: '0.75rem' }}>Charger les écritures ERP</p>
                </a>

                <a href="/validation" className="card" style={{
                    textAlign: 'center',
                    transition: 'transform 0.2s, border-color 0.2s',
                    cursor: 'pointer'
                }}>
                    <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>✅</div>
                    <h3 style={{ fontSize: '1rem', marginBottom: '0.25rem' }}>Validation</h3>
                    <p style={{ color: '#888', fontSize: '0.75rem' }}>Approuver les propositions</p>
                </a>

                <a href="/settings" className="card" style={{
                    textAlign: 'center',
                    transition: 'transform 0.2s, border-color 0.2s',
                    cursor: 'pointer'
                }}>
                    <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>⚙️</div>
                    <h3 style={{ fontSize: '1rem', marginBottom: '0.25rem' }}>Paramètres</h3>
                    <p style={{ color: '#888', fontSize: '0.75rem' }}>Configuration cabinet</p>
                </a>
            </div>

            {/* Footer */}
            <footer style={{
                marginTop: 'auto',
                color: '#666',
                fontSize: '0.75rem',
                textAlign: 'center'
            }}>
                <p>Fiducia MVP • 100% des lignes 471 justifiées</p>
            </footer>
        </main>
    );
}
