'use client';

import { useState, useRef } from 'react';

const CABINET_ID = '11111111-1111-1111-1111-111111111111'; // Test cabinet

interface Voice {
    id: string;
    name: string;
    voice_id: string;
    created_at: string;
}

export default function SettingsPage() {
    const [isRecording, setIsRecording] = useState(false);
    const [audioBlob, setAudioBlob] = useState<Blob | null>(null);
    const [audioUrl, setAudioUrl] = useState<string | null>(null);
    const [voiceName, setVoiceName] = useState('');
    const [uploading, setUploading] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
    const [voices, setVoices] = useState<Voice[]>([]);

    const mediaRecorderRef = useRef<MediaRecorder | null>(null);
    const chunksRef = useRef<Blob[]>([]);
    const fileInputRef = useRef<HTMLInputElement>(null);

    const startRecording = async () => {
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            const mediaRecorder = new MediaRecorder(stream);
            mediaRecorderRef.current = mediaRecorder;
            chunksRef.current = [];

            mediaRecorder.ondataavailable = (e) => {
                if (e.data.size > 0) {
                    chunksRef.current.push(e.data);
                }
            };

            mediaRecorder.onstop = () => {
                const blob = new Blob(chunksRef.current, { type: 'audio/webm' });
                setAudioBlob(blob);
                setAudioUrl(URL.createObjectURL(blob));
                stream.getTracks().forEach(track => track.stop());
            };

            mediaRecorder.start();
            setIsRecording(true);
        } catch (err) {
            setMessage({ type: 'error', text: 'Erreur accès microphone. Vérifiez les permissions.' });
        }
    };

    const stopRecording = () => {
        if (mediaRecorderRef.current && isRecording) {
            mediaRecorderRef.current.stop();
            setIsRecording(false);
        }
    };

    const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) {
            setAudioBlob(file);
            setAudioUrl(URL.createObjectURL(file));
        }
    };

    const clearAudio = () => {
        setAudioBlob(null);
        setAudioUrl(null);
        if (fileInputRef.current) {
            fileInputRef.current.value = '';
        }
    };

    const uploadVoice = async () => {
        if (!audioBlob || !voiceName.trim()) {
            setMessage({ type: 'error', text: 'Veuillez enregistrer un audio et entrer un nom.' });
            return;
        }

        setUploading(true);
        setMessage(null);

        try {
            const formData = new FormData();
            formData.append('audio', audioBlob, 'voice_sample.webm');
            formData.append('name', voiceName);

            // Use a test collaborator ID
            const collaboratorId = '22222222-2222-2222-2222-222222222222';

            const res = await fetch(`/api/v1/collaborators/${collaboratorId}/voice/clone`, {
                method: 'POST',
                body: formData,
            });

            if (res.ok) {
                const data = await res.json();
                setMessage({ type: 'success', text: `Voix "${data.name}" clonée avec succès ! ID: ${data.voice_id}` });
                clearAudio();
                setVoiceName('');
            } else {
                const err = await res.json();
                setMessage({ type: 'error', text: err.error || 'Échec du clonage' });
            }
        } catch (err) {
            setMessage({ type: 'error', text: 'Erreur réseau' });
        } finally {
            setUploading(false);
        }
    };

    return (
        <main style={{ minHeight: '100vh', padding: '2rem' }}>
            <h1 style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>
                ⚙️ Paramètres
            </h1>

            {/* Voice Cloning Section */}
            <div className="card" style={{ marginBottom: '1.5rem' }}>
                <h2 style={{ fontSize: '1.25rem', marginBottom: '1rem' }}>
                    🎙️ Clonage de Voix
                </h2>
                <p style={{ color: '#888', marginBottom: '1rem' }}>
                    Enregistrez un échantillon de votre voix (30 secondes minimum recommandé) pour générer des messages vocaux personnalisés.
                </p>

                {/* Recording Controls */}
                <div style={{
                    padding: '1.5rem',
                    background: 'rgba(99, 102, 241, 0.1)',
                    borderRadius: '0.5rem',
                    marginBottom: '1rem'
                }}>
                    <h3 style={{ fontSize: '1rem', marginBottom: '1rem' }}>📹 Enregistrer</h3>

                    <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', marginBottom: '1rem' }}>
                        {!isRecording ? (
                            <button
                                className="btn btn-primary"
                                onClick={startRecording}
                                disabled={!!audioBlob}
                            >
                                🎤 Démarrer l'enregistrement
                            </button>
                        ) : (
                            <button
                                className="btn btn-secondary"
                                onClick={stopRecording}
                                style={{ background: '#ef4444' }}
                            >
                                ⏹️ Arrêter ({'>'}30s recommandé)
                            </button>
                        )}

                        {isRecording && (
                            <div style={{
                                width: '12px',
                                height: '12px',
                                borderRadius: '50%',
                                background: '#ef4444',
                                animation: 'pulse 1s infinite'
                            }} />
                        )}
                    </div>

                    <p style={{ color: '#666', fontSize: '0.875rem' }}>
                        <strong>Conseils :</strong> Parlez clairement pendant au moins 30 secondes. Lisez un texte à voix haute pour un meilleur résultat.
                    </p>
                </div>

                {/* File Upload */}
                <div style={{
                    padding: '1.5rem',
                    background: 'rgba(34, 197, 94, 0.1)',
                    borderRadius: '0.5rem',
                    marginBottom: '1rem'
                }}>
                    <h3 style={{ fontSize: '1rem', marginBottom: '1rem' }}>📁 Uploader un fichier audio</h3>

                    <input
                        ref={fileInputRef}
                        type="file"
                        accept="audio/*"
                        onChange={handleFileUpload}
                        style={{ marginBottom: '0.5rem' }}
                    />
                    <p style={{ color: '#666', fontSize: '0.75rem' }}>
                        Formats acceptés : MP3, WAV, M4A, OGG (max 10MB)
                    </p>
                </div>

                {/* Audio Preview */}
                {audioUrl && (
                    <div style={{
                        padding: '1rem',
                        background: 'rgba(59, 130, 246, 0.1)',
                        borderRadius: '0.5rem',
                        marginBottom: '1rem'
                    }}>
                        <h3 style={{ fontSize: '1rem', marginBottom: '0.5rem' }}>🔊 Aperçu</h3>
                        <audio controls src={audioUrl} style={{ width: '100%' }} />
                        <button
                            onClick={clearAudio}
                            style={{ marginTop: '0.5rem', color: '#ef4444', background: 'none', border: 'none', cursor: 'pointer' }}
                        >
                            🗑️ Supprimer et recommencer
                        </button>
                    </div>
                )}

                {/* Voice Name & Submit */}
                <div style={{
                    padding: '1rem',
                    background: '#1a1a2e',
                    borderRadius: '0.5rem',
                    marginBottom: '1rem'
                }}>
                    <label style={{ display: 'block', marginBottom: '0.5rem' }}>
                        Nom de la voix *
                    </label>
                    <input
                        type="text"
                        value={voiceName}
                        onChange={(e) => setVoiceName(e.target.value)}
                        placeholder="ex: Jean Dupont - Cabinet XYZ"
                        style={{
                            width: '100%',
                            padding: '0.75rem',
                            borderRadius: '0.5rem',
                            border: '1px solid #333',
                            background: '#0a0a14',
                            color: '#fff',
                            marginBottom: '1rem'
                        }}
                    />

                    <button
                        className="btn btn-primary"
                        onClick={uploadVoice}
                        disabled={!audioBlob || !voiceName.trim() || uploading}
                        style={{ width: '100%' }}
                    >
                        {uploading ? '⏳ Clonage en cours...' : '🚀 Cloner ma voix'}
                    </button>
                </div>

                {/* Messages */}
                {message && (
                    <div style={{
                        padding: '1rem',
                        borderRadius: '0.5rem',
                        background: message.type === 'success' ? 'rgba(34, 197, 94, 0.2)' : 'rgba(239, 68, 68, 0.2)',
                        border: `1px solid ${message.type === 'success' ? '#22c55e' : '#ef4444'}`
                    }}>
                        {message.type === 'success' ? '✅' : '❌'} {message.text}
                    </div>
                )}
            </div>

            {/* How It Works */}
            <div className="card">
                <h2 style={{ fontSize: '1.25rem', marginBottom: '1rem' }}>
                    ℹ️ Comment ça marche ?
                </h2>

                <ol style={{ paddingLeft: '1.5rem', lineHeight: '1.8' }}>
                    <li><strong>Enregistrez</strong> ou uploadez un échantillon audio de votre voix (30 secondes minimum)</li>
                    <li><strong>Nommez</strong> votre voix (ex: "Marie - Comptable Senior")</li>
                    <li><strong>Cliquez</strong> sur "Cloner ma voix" - ElevenLabs analyse et crée votre clone vocal</li>
                    <li><strong>Utilisez</strong> votre voix clonée pour envoyer des relances vocales personnalisées aux clients</li>
                </ol>

                <div style={{
                    marginTop: '1rem',
                    padding: '1rem',
                    background: 'rgba(251, 191, 36, 0.1)',
                    borderRadius: '0.5rem',
                    border: '1px solid rgba(251, 191, 36, 0.3)'
                }}>
                    <strong>⚠️ Note :</strong> Le clonage de voix nécessite une clé API ElevenLabs avec les permissions appropriées.
                    La qualité du clone dépend de la qualité et durée de l'échantillon audio.
                </div>
            </div>

            <style jsx>{`
                @keyframes pulse {
                    0%, 100% { opacity: 1; }
                    50% { opacity: 0.5; }
                }
            `}</style>
        </main>
    );
}
