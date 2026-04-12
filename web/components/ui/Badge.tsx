import React from 'react';
import { cn } from '@/lib/utils';
import { cva, type VariantProps } from "class-variance-authority";

const badgeVariants = cva(
  "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors duration-fast",
  {
    variants: {
      variant: {
        default: "bg-neutral-800 text-neutral-300 border border-neutral-700",
        primary: "bg-primary-600/20 text-primary-400 border border-primary-500/30",
        success: "bg-success-600/20 text-success-400 border border-success-500/30",
        warning: "bg-warning-600/20 text-warning-400 border border-warning-500/30",
        error: "bg-error-600/20 text-error-400 border border-error-500/30",
        info: "bg-info-600/20 text-info-400 border border-info-500/30",
        outline: "border border-neutral-700 text-neutral-400 bg-transparent",
      },
      size: {
        xs: "px-1.5 py-0.5 text-[10px]",
        sm: "px-2 py-0.5 text-xs",
        default: "px-2.5 py-0.5 text-xs",
        lg: "px-3 py-1 text-sm",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {
  children: React.ReactNode;
  icon?: React.ReactNode;
}

const Badge = React.forwardRef<HTMLSpanElement, BadgeProps>(
  ({ className, variant, size, icon, children, ...props }, ref) => {
    return (
      <span
        ref={ref}
        className={cn(badgeVariants({ variant, size, className }))}
        {...props}
      >
        {icon && <span className="mr-1">{icon}</span>}
        {children}
      </span>
    );
  }
);

Badge.displayName = 'Badge';

export default Badge;
