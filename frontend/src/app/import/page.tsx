'use client';

import { useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';

interface PreviewData {
    filename: string;
    size: number;
    rows: string[][];
    total_rows: number;
    detected: {
        mapping: {
            amount_column: number;
            date_column: number;
            label_column: number;
        };
        confidence: number;
        headers: string[];
    };
}

interface ImportResult {
    batch_id: string;
    total_rows: number;
    imported_rows: number;
    failed_rows: number;
    errors: { row: number; message: string }[];
}

export default function ImportPage() {
    const router = useRouter();
    const [file, setFile] = useState<File | null>(null);
    const [preview, setPreview] = useState<PreviewData | null>(null);
    const [importing, setImporting] = useState(false);
    const [result, setResult] = useState<ImportResult | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [dragActive, setDragActive] = useState(false);

    // Column mapping state
    const [mapping, setMapping] = useState({
        amount_column: 0,
        date_column: 0,
        label_column: 0,
    });

    // Demo cabinet ID (in production, get from auth context)
    const cabinetId = '00000000-0000-0000-0000-000000000001';

    const handleDrag = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        if (e.type === 'dragenter' || e.type === 'dragover') {
            setDragActive(true);
        } else if (e.type === 'dragleave') {
            setDragActive(false);
        }
    }, []);

    const handleDrop = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setDragActive(false);

        if (e.dataTransfer.files && e.dataTransfer.files[0]) {
            handleFile(e.dataTransfer.files[0]);
        }
    }, []);

    const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files[0]) {
            handleFile(e.target.files[0]);
        }
    };

    const handleFile = async (selectedFile: File) => {
        setFile(selectedFile);
        setError(null);
        setResult(null);
        setPreview(null);

        // Upload for preview
        const formData = new FormData();
        formData.append('file', selectedFile);

        try {
            const res = await fetch(`/api/v1/cabinets/${cabinetId}/import/preview`, {
                method: 'POST',
                body: formData,
            });

            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || 'Failed to preview file');
            }

            const data: PreviewData = await res.json();
            setPreview(data);

            // Set detected mapping
            if (data.detected) {
                setMapping({
                    amount_column: data.detected.mapping.amount_column,
                    date_column: data.detected.mapping.date_column,
                    label_column: data.detected.mapping.label_column,
                });
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to preview file');
        }
    };

    const handleImport = async () => {
        if (!file) return;

        setImporting(true);
        setError(null);

        const formData = new FormData();
        formData.append('file', file);
        formData.append('mapping', JSON.stringify(mapping));

        try {
            const res = await fetch(`/api/v1/cabinets/${cabinetId}/import/csv`, {
                method: 'POST',
                body: formData,
            });

            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || 'Import failed');
            }

            const data: ImportResult = await res.json();
            setResult(data);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Import failed');
        } finally {
            setImporting(false);
        }
    };

    const resetForm = () => {
        setFile(null);
        setPreview(null);
        setResult(null);
        setError(null);
    };

    return (
        <main style={{ minHeight: '100vh', padding: '2rem' }}>
            {/* Header */}
            <div style={{ marginBottom: '2rem' }}>
                <a href="/" style={{ color: '#888', fontSize: '0.875rem' }}>← Retour</a>
                <h1 style={{ fontSize: '1.75rem', marginTop: '0.5rem' }}>
                    Import CSV
                </h1>
                <p style={{ color: '#888' }}>
                    Importez vos écritures du compte 471 depuis votre ERP
                </p>
            </div>

            {/* Result Screen */}
            {result && (
                <div className="card animate-fadeIn" style={{ maxWidth: '600px', margin: '0 auto' }}>
                    <div style={{ textAlign: 'center', marginBottom: '1.5rem' }}>
                        <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>
                            {result.failed_rows === 0 ? '✅' : '⚠️'}
                        </div>
                        <h2 style={{ fontSize: '1.25rem' }}>Import terminé</h2>
                    </div>

                    <div style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(3, 1fr)',
                        gap: '1rem',
                        marginBottom: '1.5rem'
                    }}>
                        <div style={{ textAlign: 'center' }}>
                            <div style={{ fontSize: '1.5rem', fontWeight: 600 }}>{result.total_rows}</div>
                            <div style={{ color: '#888', fontSize: '0.75rem' }}>Total lignes</div>
                        </div>
                        <div style={{ textAlign: 'center' }}>
                            <div style={{ fontSize: '1.5rem', fontWeight: 600, color: '#22c55e' }}>
                                {result.imported_rows}
                            </div>
                            <div style={{ color: '#888', fontSize: '0.75rem' }}>Importées</div>
                        </div>
                        <div style={{ textAlign: 'center' }}>
                            <div style={{ fontSize: '1.5rem', fontWeight: 600, color: result.failed_rows > 0 ? '#ef4444' : '#888' }}>
                                {result.failed_rows}
                            </div>
                            <div style={{ color: '#888', fontSize: '0.75rem' }}>Erreurs</div>
                        </div>
                    </div>

                    {result.errors && result.errors.length > 0 && (
                        <div style={{
                            background: 'rgba(239, 68, 68, 0.1)',
                            borderRadius: '0.5rem',
                            padding: '1rem',
                            marginBottom: '1.5rem',
                            maxHeight: '150px',
                            overflow: 'auto'
                        }}>
                            <div style={{ fontWeight: 500, marginBottom: '0.5rem', color: '#ef4444' }}>
                                Détail des erreurs:
                            </div>
                            {result.errors.slice(0, 10).map((err, idx) => (
                                <div key={idx} style={{ fontSize: '0.75rem', color: '#888' }}>
                                    Ligne {err.row}: {err.message}
                                </div>
                            ))}
                            {result.errors.length > 10 && (
                                <div style={{ fontSize: '0.75rem', color: '#888', marginTop: '0.5rem' }}>
                                    ...et {result.errors.length - 10} autres erreurs
                                </div>
                            )}
                        </div>
                    )}

                    <div style={{ display: 'flex', gap: '1rem' }}>
                        <button
                            className="btn btn-primary"
                            style={{ flex: 1 }}
                            onClick={() => router.push('/dashboard')}
                        >
                            Voir le Dashboard
                        </button>
                        <button
                            className="btn btn-secondary"
                            onClick={resetForm}
                        >
                            Nouvel import
                        </button>
                    </div>
                </div>
            )}

            {/* Upload / Preview Screen */}
            {!result && (
                <div style={{ maxWidth: '800px', margin: '0 auto' }}>
                    {/* Drop Zone */}
                    <div
                        onDragEnter={handleDrag}
                        onDragLeave={handleDrag}
                        onDragOver={handleDrag}
                        onDrop={handleDrop}
                        style={{
                            border: `2px dashed ${dragActive ? '#6366f1' : '#333'}`,
                            borderRadius: '0.75rem',
                            padding: '3rem',
                            textAlign: 'center',
                            background: dragActive ? 'rgba(99, 102, 241, 0.1)' : 'transparent',
                            transition: 'all 0.2s',
                            marginBottom: '1.5rem',
                            cursor: 'pointer'
                        }}
                        onClick={() => document.getElementById('file-input')?.click()}
                    >
                        <input
                            id="file-input"
                            type="file"
                            accept=".csv,.txt"
                            onChange={handleFileInput}
                            style={{ display: 'none' }}
                        />
                        <div style={{ fontSize: '2.5rem', marginBottom: '1rem' }}>📄</div>
                        <p style={{ marginBottom: '0.5rem' }}>
                            {file ? file.name : 'Glissez votre fichier CSV ici'}
                        </p>
                        <p style={{ color: '#888', fontSize: '0.875rem' }}>
                            ou cliquez pour sélectionner
                        </p>
                        {file && (
                            <p style={{ color: '#888', fontSize: '0.75rem', marginTop: '0.5rem' }}>
                                {(file.size / 1024).toFixed(1)} KB
                            </p>
                        )}
                    </div>

                    {/* Error */}
                    {error && (
                        <div className="badge badge-error" style={{
                            display: 'block',
                            padding: '1rem',
                            marginBottom: '1.5rem',
                            borderRadius: '0.5rem'
                        }}>
                            {error}
                        </div>
                    )}

                    {/* Preview */}
                    {preview && (
                        <div className="card animate-fadeIn">
                            <div style={{
                                display: 'flex',
                                justifyContent: 'space-between',
                                alignItems: 'center',
                                marginBottom: '1rem'
                            }}>
                                <h3 style={{ fontSize: '1rem' }}>Aperçu ({preview.total_rows} lignes)</h3>
                                {preview.detected && (
                                    <span className={`badge ${preview.detected.confidence > 0.7 ? 'badge-success' : 'badge-pending'}`}>
                                        Confiance: {Math.round(preview.detected.confidence * 100)}%
                                    </span>
                                )}
                            </div>

                            {/* Column Mapping */}
                            <div style={{
                                display: 'grid',
                                gridTemplateColumns: 'repeat(3, 1fr)',
                                gap: '1rem',
                                marginBottom: '1.5rem',
                                padding: '1rem',
                                background: '#0a0a0a',
                                borderRadius: '0.5rem'
                            }}>
                                <div>
                                    <label style={{ fontSize: '0.75rem', color: '#888', display: 'block', marginBottom: '0.25rem' }}>
                                        Colonne Montant
                                    </label>
                                    <select
                                        className="input"
                                        value={mapping.amount_column}
                                        onChange={(e) => setMapping({ ...mapping, amount_column: parseInt(e.target.value) })}
                                    >
                                        {preview.detected.headers.map((h, i) => (
                                            <option key={i} value={i}>{h || `Colonne ${i + 1}`}</option>
                                        ))}
                                    </select>
                                </div>
                                <div>
                                    <label style={{ fontSize: '0.75rem', color: '#888', display: 'block', marginBottom: '0.25rem' }}>
                                        Colonne Date
                                    </label>
                                    <select
                                        className="input"
                                        value={mapping.date_column}
                                        onChange={(e) => setMapping({ ...mapping, date_column: parseInt(e.target.value) })}
                                    >
                                        {preview.detected.headers.map((h, i) => (
                                            <option key={i} value={i}>{h || `Colonne ${i + 1}`}</option>
                                        ))}
                                    </select>
                                </div>
                                <div>
                                    <label style={{ fontSize: '0.75rem', color: '#888', display: 'block', marginBottom: '0.25rem' }}>
                                        Colonne Libellé
                                    </label>
                                    <select
                                        className="input"
                                        value={mapping.label_column}
                                        onChange={(e) => setMapping({ ...mapping, label_column: parseInt(e.target.value) })}
                                    >
                                        {preview.detected.headers.map((h, i) => (
                                            <option key={i} value={i}>{h || `Colonne ${i + 1}`}</option>
                                        ))}
                                    </select>
                                </div>
                            </div>

                            {/* Data Preview Table */}
                            <div style={{ overflowX: 'auto' }}>
                                <table style={{ fontSize: '0.75rem' }}>
                                    <thead>
                                        <tr>
                                            {preview.rows[0]?.map((header, i) => (
                                                <th key={i} style={{
                                                    background:
                                                        i === mapping.amount_column ? 'rgba(99, 102, 241, 0.2)' :
                                                            i === mapping.date_column ? 'rgba(34, 197, 94, 0.2)' :
                                                                i === mapping.label_column ? 'rgba(245, 158, 11, 0.2)' :
                                                                    'transparent'
                                                }}>
                                                    {header || `Col ${i + 1}`}
                                                </th>
                                            ))}
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {preview.rows.slice(1, 6).map((row, i) => (
                                            <tr key={i}>
                                                {row.map((cell, j) => (
                                                    <td key={j} style={{
                                                        background:
                                                            j === mapping.amount_column ? 'rgba(99, 102, 241, 0.1)' :
                                                                j === mapping.date_column ? 'rgba(34, 197, 94, 0.1)' :
                                                                    j === mapping.label_column ? 'rgba(245, 158, 11, 0.1)' :
                                                                        'transparent'
                                                    }}>
                                                        {cell}
                                                    </td>
                                                ))}
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>

                            {preview.rows.length > 6 && (
                                <p style={{ color: '#888', fontSize: '0.75rem', marginTop: '0.5rem', textAlign: 'center' }}>
                                    ...et {preview.total_rows - 5} autres lignes
                                </p>
                            )}

                            {/* Import Button */}
                            <div style={{ marginTop: '1.5rem', display: 'flex', gap: '1rem' }}>
                                <button
                                    className="btn btn-primary"
                                    style={{ flex: 1 }}
                                    onClick={handleImport}
                                    disabled={importing}
                                >
                                    {importing ? (
                                        <span className="animate-pulse">Import en cours...</span>
                                    ) : (
                                        `Importer ${preview.total_rows} lignes`
                                    )}
                                </button>
                                <button
                                    className="btn btn-secondary"
                                    onClick={resetForm}
                                    disabled={importing}
                                >
                                    Annuler
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            )}
        </main>
    );
}
