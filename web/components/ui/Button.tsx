import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"
import { motion } from "framer-motion"
import { MaterialIcon } from "@/components/ui/MaterialIcon"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "btn-base",
  {
    variants: {
      variant: {
        default: "bg-primary-600 text-white hover:bg-primary-700 focus-visible:ring-primary-500",
        destructive: "bg-error-600 text-white hover:bg-error-700 focus-visible:ring-error-500",
        outline: "border border-neutral-700 bg-neutral-900/50 hover:bg-neutral-800/50 text-neutral-50 focus-visible:ring-primary-500",
        secondary: "bg-neutral-800 text-neutral-300 hover:bg-neutral-700 focus-visible:ring-neutral-500",
        ghost: "hover:bg-neutral-800/50 text-neutral-400 hover:text-neutral-300 focus-visible:ring-neutral-500",
        link: "text-primary-600 underline-offset-4 hover:underline focus-visible:ring-primary-500",
        success: "bg-success-600 text-white hover:bg-success-700 focus-visible:ring-success-500",
        warning: "bg-warning-600 text-white hover:bg-warning-700 focus-visible:ring-warning-500",
        info: "bg-info-600 text-white hover:bg-info-700 focus-visible:ring-info-500",
      },
      size: {
        xs: "h-8 rounded-md px-2 text-xs",
        sm: "h-9 rounded-lg px-3 text-sm",
        default: "h-10 rounded-lg px-4 text-sm",
        lg: "h-11 rounded-lg px-6 text-base",
        xl: "h-12 rounded-xl px-8 text-lg",
        icon: "h-10 w-10 rounded-lg",
        "icon-sm": "h-8 w-8 rounded-md",
        "icon-lg": "h-12 w-12 rounded-xl",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
  loading?: boolean
  icon?: React.ReactNode
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, loading = false, icon, children, disabled, ...props }, ref) => {
    if (asChild) {
      return (
        <Slot
          className={cn(buttonVariants({ variant, size, className }))}
          ref={ref}
          {...props}
        >
          {children}
        </Slot>
      )
    }

    // Filter out motion-specific props that don't belong on button
    const { 
      onAnimationStart, 
      onAnimationEnd, 
      whileHover, 
      whileTap, 
      transition, 
      initial, 
      animate, 
      exit, 
      variants, 
      ...buttonProps 
    } = props as any
    
    return (
      <motion.button
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        disabled={disabled || loading}
        whileHover={{ scale: 1.02 }}
        whileTap={{ scale: 0.98 }}
        transition={{ type: "spring", stiffness: 400, damping: 17 }}
        {...buttonProps}
      >
        {loading && (
          <MaterialIcon name="sync" className="mr-2 h-4 w-4 animate-spin" />
        )}
        {icon && !loading && (
          <span className="mr-2 flex items-center">{icon}</span>
        )}
        {children}
      </motion.button>
    )
  }
)
Button.displayName = "Button"

export { Button, buttonVariants }
export default Button
