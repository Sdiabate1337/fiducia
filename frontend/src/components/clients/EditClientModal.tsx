import { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useAuth } from '@/context/AuthContext';
import { toast } from 'sonner';

interface Client {
    id: string;
    name: string;
    email?: string;
    phone?: string;
}

interface EditClientModalProps {
    isOpen: boolean;
    onClose: () => void;
    client: Client | null;
    onUpdated: (client: Client) => void;
}

export function EditClientModal({ isOpen, onClose, client, onUpdated }: EditClientModalProps) {
    const { token } = useAuth();
    const [email, setEmail] = useState('');
    const [phone, setPhone] = useState('');
    const [isLoading, setIsLoading] = useState(false);

    useEffect(() => {
        if (client) {
            setEmail(client.email || '');
            setPhone(client.phone || '');
        }
    }, [client]);

    const handleSave = async () => {
        if (!client) return;
        setIsLoading(true);
        try {
            const res = await fetch(`http://localhost:8080/api/v1/clients/${client.id}`, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ email, phone })
            });

            if (!res.ok) throw new Error("Failed to update client");

            const updated = await res.json();
            toast.success("Contact mis à jour");
            onUpdated(updated);
            onClose();
        } catch (e) {
            toast.error("Erreur lors de la mise à jour");
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle>Modifier le Contact</DialogTitle>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="name" className="text-right">Nom</Label>
                        <Input id="name" value={client?.name || ''} disabled className="col-span-3" />
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="email" className="text-right">Email</Label>
                        <Input id="email" value={email} onChange={(e) => setEmail(e.target.value)} className="col-span-3" />
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="phone" className="text-right">Tel</Label>
                        <Input id="phone" value={phone} onChange={(e) => setPhone(e.target.value)} className="col-span-3" />
                    </div>
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={onClose} disabled={isLoading}>Annuler</Button>
                    <Button onClick={handleSave} disabled={isLoading} className="bg-[#1A4D2E] hover:bg-[#1A4D2E]/90 text-white shadow-lg shadow-[#1A4D2E]/20">Enregistrer</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
