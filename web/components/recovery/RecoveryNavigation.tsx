'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { motion } from 'framer-motion'
import { cn } from '@/lib/utils'

const recoveryNavItems = [
  { name: 'Overview', href: '/recovery' },
  { name: 'Jobs', href: '/recovery/jobs' },
  { name: 'Workflows', href: '/recovery/workflows' },
  { name: 'Templates', href: '/recovery/templates' },
  { name: 'Performance', href: '/recovery/performance' },
]

export function RecoveryNavigation() {
  const pathname = usePathname()

  return (
    <nav className="flex space-x-1 bg-muted p-1 rounded-lg">
      {recoveryNavItems.map((item) => {
        const isActive = pathname === item.href
        
        return (
          <Link key={item.name} href={item.href}>
            <motion.div
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              className={cn(
                'relative px-3 py-2 text-sm font-medium rounded-md transition-colors',
                isActive
                  ? 'text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              {isActive && (
                <motion.div
                  layoutId="recovery-nav-active"
                  className="absolute inset-0 bg-background rounded-md shadow-sm"
                  transition={{ type: 'spring', stiffness: 380, damping: 30 }}
                />
              )}
              <span className="relative z-10">{item.name}</span>
            </motion.div>
          </Link>
        )
      })}
    </nav>
  )
}
