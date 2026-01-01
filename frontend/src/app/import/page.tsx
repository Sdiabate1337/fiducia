'use client';

import { useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { motion, AnimatePresence } from 'framer-motion';
import {
    Upload, FileText, CheckCircle2, AlertTriangle,
    ArrowLeft, ArrowRight, Database, LayoutTemplate,
    Calendar, CreditCard, Type, X
} from 'lucide-react';
import { cn } from '@/lib/utils';
import Link from 'next/link';

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
        <div className="min-h-screen bg-[#F9F8F6] text-[#1A1A1A] font-sans selection:bg-[#4F2830]/20 flex flex-col">
            {/* Navbar */}
            <nav className="shrink-0 bg-[#F9F8F6] border-b border-[#1A1A1A]/5 px-4 md:px-8 h-16 flex items-center justify-between z-50 sticky top-0 backdrop-blur-md bg-opacity-90">
                <div className="flex items-center gap-4 md:gap-6">
                    <Link href="/dashboard" className="p-2 -ml-2 text-[#1A1A1A]/40 hover:text-[#1A1A1A] transition-colors rounded-lg hover:bg-[#1A1A1A]/5">
                        <ArrowLeft size={20} />
                    </Link>
                    <span className="text-xl font-serif font-bold tracking-tight">Import des Écritures</span>
                </div>
            </nav>

            <main className="flex-1 max-w-5xl mx-auto w-full p-4 md:p-8">
                <AnimatePresence mode="wait">
                    {/* 1. Success State */}
                    {result ? (
                        <motion.div
                            key="result"
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -20 }}
                            className="bg-white rounded-3xl p-8 md:p-12 border border-[#1A1A1A]/5 shadow-xl text-center max-w-2xl mx-auto"
                        >
                            <div className="w-20 h-20 rounded-full bg-[#1A4D2E]/5 text-[#1A4D2E] flex items-center justify-center mx-auto mb-6">
                                {result.failed_rows === 0 ? <CheckCircle2 size={40} /> : <AlertTriangle size={40} className="text-amber-500" />}
                            </div>

                            <h2 className="text-3xl md:text-4xl font-serif font-bold mb-2">Import Terminé</h2>
                            <p className="text-[#1A1A1A]/60 font-medium mb-10">
                                {result.failed_rows === 0
                                    ? "Toutes les lignes ont été intégrées avec succès."
                                    : "L'import est terminé avec quelques avertissements."}
                            </p>

                            <div className="grid grid-cols-3 gap-4 mb-10">
                                <div className="p-4 rounded-2xl bg-[#F9F8F6] border border-[#1A1A1A]/5">
                                    <div className="text-2xl font-bold font-serif mb-1">{result.total_rows}</div>
                                    <div className="text-[10px] uppercase font-bold tracking-widest text-[#1A1A1A]/40">Total</div>
                                </div>
                                <div className="p-4 rounded-2xl bg-[#F9F8F6] border border-[#1A1A1A]/5">
                                    <div className="text-2xl font-bold font-serif mb-1 text-[#1A4D2E]">{result.imported_rows}</div>
                                    <div className="text-[10px] uppercase font-bold tracking-widest text-[#1A1A1A]/40">Succès</div>
                                </div>
                                <div className="p-4 rounded-2xl bg-[#F9F8F6] border border-[#1A1A1A]/5">
                                    <div className={`text-2xl font-bold font-serif mb-1 ${result.failed_rows > 0 ? 'text-red-500' : 'text-[#1A1A1A]/40'}`}>{result.failed_rows}</div>
                                    <div className="text-[10px] uppercase font-bold tracking-widest text-[#1A1A1A]/40">Erreurs</div>
                                </div>
                            </div>

                            {result.errors && result.errors.length > 0 && (
                                <div className="text-left mb-8 p-4 bg-red-50 rounded-xl border border-red-100 max-h-40 overflow-y-auto custom-scrollbar">
                                    <h4 className="text-xs font-bold uppercase text-red-800 mb-2 sticky top-0 bg-red-50 pb-2">Rapport d'erreurs</h4>
                                    <ul className="space-y-1">
                                        {result.errors.map((err, idx) => (
                                            <li key={idx} className="text-xs text-red-600 font-mono">
                                                Ligne {err.row}: {err.message}
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            )}

                            <div className="flex gap-4 justify-center">
                                <Link href="/dashboard" className="px-8 py-3 bg-[#1A1A1A] text-white rounded-xl font-medium hover:bg-[#1A4D2E] transition-all shadow-lg hover:shadow-xl hover:-translate-y-0.5">
                                    Voir le Dashboard
                                </Link>
                                <button onClick={resetForm} className="px-8 py-3 bg-white border border-[#1A1A1A]/10 text-[#1A1A1A] rounded-xl font-medium hover:bg-[#F9F8F6] transition-all">
                                    Nouvel Import
                                </button>
                            </div>
                        </motion.div>
                    ) : (
                        /* 2. Upload / Preview State */
                        <motion.div
                            key="upload"
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            className="space-y-8"
                        >
                            {/* Upload Zone */}
                            {!preview && (
                                <div
                                    onDragEnter={handleDrag}
                                    onDragLeave={handleDrag}
                                    onDragOver={handleDrag}
                                    onDrop={handleDrop}
                                    onClick={() => document.getElementById('file-input')?.click()}
                                    className={cn(
                                        "group border-2 border-dashed rounded-3xl p-12 md:p-20 text-center transition-all duration-300 cursor-pointer relative overflow-hidden",
                                        dragActive
                                            ? "border-[#1A4D2E] bg-[#1A4D2E]/5"
                                            : "border-[#1A1A1A]/10 hover:border-[#1A1A1A]/30 hover:bg-white"
                                    )}
                                >
                                    <input
                                        id="file-input"
                                        type="file"
                                        accept=".csv,.txt"
                                        onChange={handleFileInput}
                                        className="hidden"
                                    />

                                    <div className="relative z-10 flex flex-col items-center">
                                        <div className={cn(
                                            "w-16 h-16 md:w-20 md:h-20 rounded-full flex items-center justify-center mb-6 transition-all duration-300",
                                            dragActive ? "bg-[#1A4D2E] text-white" : "bg-[#1A1A1A]/5 text-[#1A1A1A]/40 group-hover:bg-[#1A1A1A] group-hover:text-white"
                                        )}>
                                            <Upload className="w-8 h-8 md:w-10 md:h-10" />
                                        </div>
                                        <h3 className="text-xl md:text-2xl font-serif font-bold mb-2">
                                            {dragActive ? "Déposez le fichier ici" : "Importez votre CSV"}
                                        </h3>
                                        <p className="text-[#1A1A1A]/50 max-w-sm mx-auto mb-8">
                                            Glissez-déposez votre export comptable (FEC) ou cliquez pour parcourir vos fichiers.
                                        </p>
                                        <span className="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-[#1A1A1A]/10 text-xs font-bold uppercase tracking-wider text-[#1A1A1A]/60 bg-white">
                                            <Database size={14} />
                                            Format supporté : CSV UTF-8
                                        </span>
                                    </div>
                                </div>
                            )}

                            {/* Error Message */}
                            <AnimatePresence>
                                {error && (
                                    <motion.div
                                        initial={{ opacity: 0, height: 0 }}
                                        animate={{ opacity: 1, height: 'auto' }}
                                        exit={{ opacity: 0, height: 0 }}
                                        className="bg-red-50 border border-red-100 text-red-800 px-6 py-4 rounded-xl flex items-center gap-3"
                                    >
                                        <AlertTriangle size={20} />
                                        {error}
                                    </motion.div>
                                )}
                            </AnimatePresence>

                            {/* Preview Section */}
                            {preview && (
                                <motion.div
                                    initial={{ opacity: 0, y: 40 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    className="bg-white rounded-3xl border border-[#1A1A1A]/5 shadow-xl overflow-hidden"
                                >
                                    {/* Header */}
                                    <div className="p-6 md:p-8 border-b border-[#1A1A1A]/5 flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                                        <div>
                                            <div className="flex items-center gap-2 mb-1">
                                                <FileText size={16} className="text-[#1A1A1A]/40" />
                                                <h3 className="font-bold text-lg">{file?.name}</h3>
                                            </div>
                                            <p className="text-[#1A1A1A]/40 text-sm">{preview.total_rows} écritures détectées</p>
                                        </div>

                                        <div className="flex items-center gap-3">
                                            <button onClick={resetForm} className="p-2 text-[#1A1A1A]/40 hover:text-red-500 transition-colors">
                                                <X size={20} />
                                            </button>
                                        </div>
                                    </div>

                                    {/* Column Mapping Grid */}
                                    <div className="p-6 md:p-8 bg-[#F9F8F6]/50 border-b border-[#1A1A1A]/5 grid md:grid-cols-3 gap-6">
                                        <div className="space-y-2">
                                            <label className="text-xs font-bold uppercase tracking-widest text-[#1A1A1A]/40 flex items-center gap-2">
                                                <CreditCard size={14} /> Montant
                                            </label>
                                            <select
                                                value={mapping.amount_column}
                                                onChange={(e) => setMapping({ ...mapping, amount_column: parseInt(e.target.value) })}
                                                className="w-full bg-white border border-[#1A1A1A]/10 rounded-xl px-4 py-3 font-medium outline-none focus:ring-2 focus:ring-[#1A1A1A]/10 transition-all hover:border-[#1A1A1A]/30"
                                            >
                                                {preview.detected.headers.map((h, i) => (
                                                    <option key={i} value={i}>{h || `Colonne ${i + 1}`}</option>
                                                ))}
                                            </select>
                                        </div>
                                        <div className="space-y-2">
                                            <label className="text-xs font-bold uppercase tracking-widest text-[#1A1A1A]/40 flex items-center gap-2">
                                                <Calendar size={14} /> Date
                                            </label>
                                            <select
                                                value={mapping.date_column}
                                                onChange={(e) => setMapping({ ...mapping, date_column: parseInt(e.target.value) })}
                                                className="w-full bg-white border border-[#1A1A1A]/10 rounded-xl px-4 py-3 font-medium outline-none focus:ring-2 focus:ring-[#1A1A1A]/10 transition-all hover:border-[#1A1A1A]/30"
                                            >
                                                {preview.detected.headers.map((h, i) => (
                                                    <option key={i} value={i}>{h || `Colonne ${i + 1}`}</option>
                                                ))}
                                            </select>
                                        </div>
                                        <div className="space-y-2">
                                            <label className="text-xs font-bold uppercase tracking-widest text-[#1A1A1A]/40 flex items-center gap-2">
                                                <Type size={14} /> Libellé
                                            </label>
                                            <select
                                                value={mapping.label_column}
                                                onChange={(e) => setMapping({ ...mapping, label_column: parseInt(e.target.value) })}
                                                className="w-full bg-white border border-[#1A1A1A]/10 rounded-xl px-4 py-3 font-medium outline-none focus:ring-2 focus:ring-[#1A1A1A]/10 transition-all hover:border-[#1A1A1A]/30"
                                            >
                                                {preview.detected.headers.map((h, i) => (
                                                    <option key={i} value={i}>{h || `Colonne ${i + 1}`}</option>
                                                ))}
                                            </select>
                                        </div>
                                    </div>

                                    {/* Data Table */}
                                    <div className="overflow-x-auto">
                                        <table className="w-full text-sm">
                                            <thead className="bg-[#F9F8F6]">
                                                <tr>
                                                    {preview.rows[0]?.slice(0, 8).map((header, i) => (
                                                        <th key={i} className={cn(
                                                            "px-6 py-4 text-left font-serif font-bold text-[#1A1A1A]",
                                                            i === mapping.amount_column && "bg-[#1A4D2E]/10 text-[#1A4D2E]",
                                                            i === mapping.date_column && "bg-[#1A4D2E]/10 text-[#1A4D2E]",
                                                            i === mapping.label_column && "bg-[#1A4D2E]/10 text-[#1A4D2E]"
                                                        )}>
                                                            {header || `Col ${i + 1}`}
                                                        </th>
                                                    ))}
                                                </tr>
                                            </thead>
                                            <tbody className="divide-y divide-[#1A1A1A]/5">
                                                {preview.rows.slice(1, 6).map((row, i) => (
                                                    <tr key={i} className="hover:bg-[#F9F8F6]/50 transition-colors">
                                                        {row.slice(0, 8).map((cell, j) => (
                                                            <td key={j} className={cn(
                                                                "px-6 py-4 font-medium text-[#1A1A1A]/80 whitespace-nowrap",
                                                                j === mapping.amount_column && "bg-[#1A4D2E]/5 font-bold text-[#1A4D2E]",
                                                                j === mapping.date_column && "bg-[#1A4D2E]/5 text-[#1A4D2E]",
                                                                j === mapping.label_column && "bg-[#1A4D2E]/5 text-[#1A4D2E]"
                                                            )}>
                                                                {cell}
                                                            </td>
                                                        ))}
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>

                                    {/* Action Footer */}
                                    <div className="p-6 md:p-8 bg-[#F9F8F6]/30 border-t border-[#1A1A1A]/5 flex justify-end">
                                        <button
                                            onClick={handleImport}
                                            disabled={importing}
                                            className="px-8 py-4 bg-[#1A1A1A] text-white rounded-xl font-medium hover:bg-[#1A4D2E] transition-all shadow-lg hover:shadow-xl hover:-translate-y-0.5 disabled:opacity-70 disabled:hover:translate-y-0 flex items-center gap-3"
                                        >
                                            {importing ? (
                                                <>
                                                    <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                                    Traitement en cours...
                                                </>
                                            ) : (
                                                <>
                                                    Confirmer l'importation <ArrowRight size={20} />
                                                </>
                                            )}
                                        </button>
                                    </div>
                                </motion.div>
                            )}
                        </motion.div>
                    )}
                </AnimatePresence>
            </main>
        </div>
    );
}
