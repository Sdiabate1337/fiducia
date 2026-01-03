import type { Metadata } from 'next';
import { Inter, Plus_Jakarta_Sans, Playfair_Display } from 'next/font/google';
import './globals.css';

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' });
const jakarta = Plus_Jakarta_Sans({ subsets: ['latin'], variable: '--font-jakarta' });
const playfair = Playfair_Display({ subsets: ['latin'], variable: '--font-playfair' });

export const metadata: Metadata = {
    title: 'Fiducia - Infrastructure Autonome',
    description: 'Comptabilité automatisée par IA',
};

import { AuthProvider } from '@/context/AuthContext';

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="fr" className={`${jakarta.variable} ${playfair.variable} ${inter.variable}`}>
            <body className="font-sans antialiased bg-[#F9F8F6]">
                <AuthProvider>
                    {children}
                </AuthProvider>
            </body>
        </html>
    );
}
