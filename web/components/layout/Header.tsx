'use client'

import { Bell, Search, User, Moon, Sun } from 'lucide-react'
import { useTheme } from 'next-themes'
import { Button, Input } from '@/components/ui'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui'

import { cn } from "@/lib/utils"

interface HeaderProps {
  className?: string;
}

import { useAppStore } from '@/lib/store'
import { MaterialIcon } from '@/components/ui'

export function Header({ className }: HeaderProps) {
  const { theme, setTheme } = useTheme()
  const { isPrivacyMode, togglePrivacyMode } = useAppStore()

  return (
    <header className={cn("flex items-center justify-between px-6 py-4 border-b border-slate-800 bg-slate-900/95 backdrop-blur supports-[backdrop-filter]:bg-slate-900/60 z-30", className)}>
      {/* Search */}
      <div className="flex items-center space-x-4 flex-1 max-w-md">
        <div className="relative group">
          <MaterialIcon name="search" className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-500 group-focus-within:text-blue-400 transition-colors" />
          <Input
            placeholder="Search recovery records, customers..."
            className="pl-10 w-80 bg-slate-800/50 border-slate-700 text-slate-100 placeholder:text-slate-500 focus-visible:ring-blue-500"
          />
        </div>
      </div>

      {/* Actions */}
      <div className="flex items-center space-x-4">
        {/* Privacy Mode Toggle */}
        <Button
          variant={isPrivacyMode ? "default" : "ghost"}
          size="sm"
          onClick={togglePrivacyMode}
          className={cn(
            "flex items-center gap-2 font-bold text-[10px] uppercase tracking-widest transition-all",
            isPrivacyMode 
              ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 hover:bg-emerald-500/30" 
              : "text-slate-400 hover:text-slate-100"
          )}
        >
          <MaterialIcon name={isPrivacyMode ? "shield" : "shield_outline"} className="text-lg" />
          <span>{isPrivacyMode ? "Privacy ON" : "Privacy OFF"}</span>
        </Button>

        {/* Theme Toggle */}
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
          className="text-slate-400 hover:text-slate-100"
        >
          <MaterialIcon name={theme === 'dark' ? 'light_mode' : 'dark_mode'} className="text-lg" />
        </Button>

        {/* Notifications */}
        <Button variant="ghost" size="sm" className="relative text-slate-400 hover:text-slate-100">
          <MaterialIcon name="notifications" className="text-lg" />
          <span className="absolute top-1 right-1 h-2 w-2 bg-blue-500 rounded-full"></span>
        </Button>

        {/* User Menu */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="flex items-center gap-2 text-slate-100 border border-slate-700 bg-slate-800/50 px-3">
              <MaterialIcon name="account_circle" className="text-xl" />
              <span className="text-xs font-bold uppercase tracking-tighter">Admin</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="bg-slate-900 border-slate-800 text-slate-200">
            <DropdownMenuItem className="hover:bg-slate-800 uppercase text-[10px] font-bold tracking-widest">
                <MaterialIcon name="person" className="mr-2 text-sm" /> Profile
            </DropdownMenuItem>
            <DropdownMenuItem className="hover:bg-slate-800 uppercase text-[10px] font-bold tracking-widest">
                <MaterialIcon name="settings" className="mr-2 text-sm" /> Settings
            </DropdownMenuItem>
            <DropdownMenuItem className="text-rose-400 hover:bg-rose-400/10 uppercase text-[10px] font-bold tracking-widest">
                <MaterialIcon name="logout" className="mr-2 text-sm" /> Logout
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}

