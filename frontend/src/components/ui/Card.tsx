import React from 'react';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
    hoverEffect?: boolean;
}

export const Card = ({ className, hoverEffect = false, children, ...props }: CardProps) => {
    const Component = hoverEffect ? motion.div : 'div';

    const motionProps = hoverEffect ? {
        whileHover: { y: -5, transition: { duration: 0.2 } },
        initial: { opacity: 0, y: 20 },
        animate: { opacity: 1, y: 0 },
        transition: { duration: 0.3 }
    } : {};

    return (
        // @ts-ignore
        <Component
            className={cn(
                "glass-card rounded-xl p-6 transition-colors",
                className
            )}
            {...motionProps}
            {...props}
        >
            {children}
        </Component>
    );
};

export const CardHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
    <div className={cn("flex flex-col space-y-1.5 mb-4", className)} {...props} />
);

export const CardTitle = ({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) => (
    <h3 className="font-sans font-semibold leading-none tracking-tight text-xl text-foreground" {...props} />
);

export const CardContent = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
    <div className={cn("pt-0", className)} {...props} />
);
