import { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useAuth } from '@/context/AuthContext';
import { toast } from 'sonner';
import { Search, UserPlus } from 'lucide-react';
import { cn } from '@/lib/utils';

interface Client {
    id: string;
    name: string;
    email?: string;
    phone?: string;
}

interface AssignClientModalProps {
    isOpen: boolean;
    onClose: () => void;
    pendingLineId: string;
    currentClientId?: string;
    onAssigned: (client: Client) => void;
    userCabinetId: string;
}

export function AssignClientModal({ isOpen, onClose, pendingLineId, currentClientId, onAssigned, userCabinetId }: AssignClientModalProps) {
    const { token } = useAuth();
    const [search, setSearch] = useState('');
    const [clients, setClients] = useState<Client[]>([]);
    const [selectedClientId, setSelectedClientId] = useState<string | null>(currentClientId || null);
    const [loading, setLoading] = useState(false);
    const [creating, setCreating] = useState(false);

    // New Client Form State
    const [newClientName, setNewClientName] = useState('');
    const [newClientEmail, setNewClientEmail] = useState('');
    const [newClientPhone, setNewClientPhone] = useState('');

    useEffect(() => {
        if (isOpen) {
            fetchClients();
        }
    }, [isOpen, search]);

    const fetchClients = async () => {
        setLoading(true);
        try {
            const query = new URLSearchParams();
            if (search) query.append('search', search);
            query.append('limit', '50');

            const res = await fetch(`http://localhost:8080/api/v1/cabinets/${userCabinetId}/clients?${query.toString()}`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (res.ok) {
                const data = await res.json();
                setClients(data.items || []);
            }
        } catch (e) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    const handleAssign = async () => {
        if (!selectedClientId) return;
        try {
            const res = await fetch(`http://localhost:8080/api/v1/pending-lines/${pendingLineId}`, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ client_id: selectedClientId })
            });

            if (!res.ok) throw new Error("Failed to assign");

            const selected = clients.find(c => c.id === selectedClientId);
            if (selected) {
                onAssigned(selected);
                toast.success(`Assigné à ${selected.name}`);
                onClose();
            }
        } catch (e) {
            toast.error("Erreur lors de l'assignation");
        }
    };

    const handleCreateClient = async () => {
        if (!newClientName) return;
        try {
            const res = await fetch(`http://localhost:8080/api/v1/cabinets/${userCabinetId}/clients`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    name: newClientName,
                    email: newClientEmail || undefined,
                    phone: newClientPhone || undefined
                })
            });

            if (!res.ok) throw new Error("Failed to create");
            const newClient = await res.json();

            // Auto Select & Refresh
            setClients([newClient, ...clients]);
            setSelectedClientId(newClient.id);
            setCreating(false);
            setNewClientName('');
            setNewClientEmail('');
            setNewClientPhone('');
            toast.success("Client créé");
        } catch (e) {
            toast.error("Erreur de création");
        }
    };

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle>Assigner un Client</DialogTitle>
                </DialogHeader>

                <div className="py-4 space-y-4">
                    {!creating ? (
                        <>
                            <div className="relative">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground w-4 h-4" />
                                <Input
                                    placeholder="Rechercher un client..."
                                    value={search}
                                    onChange={(e) => setSearch(e.target.value)}
                                    className="pl-9"
                                />
                            </div>

                            <div className="max-h-[200px] overflow-y-auto space-y-2">
                                {clients.map(client => (
                                    <div
                                        key={client.id}
                                        onClick={() => setSelectedClientId(client.id)}
                                        className={cn(
                                            "p-3 rounded-lg border cursor-pointer hover:bg-slate-50 transition-colors flex justify-between items-center",
                                            selectedClientId === client.id ? "border-black bg-slate-50 ring-1 ring-black" : "border-transparent bg-white shadow-sm"
                                        )}
                                    >
                                        <div>
                                            <div className="font-medium">{client.name}</div>
                                            {client.phone && <div className="text-xs text-muted-foreground">{client.phone}</div>}
                                        </div>
                                        {selectedClientId === client.id && <div className="w-3 h-3 bg-black rounded-full" />}
                                    </div>
                                ))}
                                {clients.length === 0 && !loading && (
                                    <div className="text-center text-sm text-muted-foreground py-4">Aucun client trouvé</div>
                                )}
                            </div>

                            <Button variant="ghost" className="w-full text-xs" onClick={() => setCreating(true)}>
                                <UserPlus className="w-4 h-4 mr-2" /> Créer un nouveau client
                            </Button>
                        </>
                    ) : (
                        <div className="space-y-4 border p-4 rounded-lg bg-slate-50">
                            <h4 className="text-sm font-medium">Nouveau Client</h4>
                            <div className="grid gap-2">
                                <Label htmlFor="new-name">Nom de l'entreprise / Client <span className="text-red-500">*</span></Label>
                                <Input
                                    id="new-name"
                                    value={newClientName}
                                    onChange={(e) => setNewClientName(e.target.value)}
                                    placeholder="Ex: Dupent SARL"
                                />
                            </div>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="grid gap-2">
                                    <Label htmlFor="new-phone">Téléphone</Label>
                                    <Input
                                        id="new-phone"
                                        value={newClientPhone}
                                        onChange={(e) => setNewClientPhone(e.target.value)}
                                        placeholder="Ex: 06..."
                                    />
                                </div>
                                <div className="grid gap-2">
                                    <Label htmlFor="new-email">Email</Label>
                                    <Input
                                        id="new-email"
                                        value={newClientEmail}
                                        onChange={(e) => setNewClientEmail(e.target.value)}
                                        placeholder="Ex: contact@..."
                                        type="email"
                                    />
                                </div>
                            </div>
                            <div className="flex justify-end gap-2">
                                <Button size="sm" variant="ghost" onClick={() => setCreating(false)}>Annuler</Button>
                                <Button
                                    size="sm"
                                    onClick={handleCreateClient}
                                    disabled={!newClientName}
                                    className="bg-[#1A4D2E] hover:bg-[#1A4D2E]/90 text-white shadow-lg shadow-[#1A4D2E]/20"
                                >
                                    Créer
                                </Button>
                            </div>
                        </div>
                    )}
                </div>

                <DialogFooter>
                    <Button variant="outline" onClick={onClose}>Annuler</Button>
                    <Button
                        onClick={handleAssign}
                        disabled={!selectedClientId || creating}
                        className="bg-[#1A4D2E] hover:bg-[#1A4D2E]/90 text-white shadow-lg shadow-[#1A4D2E]/20"
                    >
                        Confirmer
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
