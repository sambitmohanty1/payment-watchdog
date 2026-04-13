'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/hooks/useAuth'
import { Button, Input, MaterialIcon } from '@/components/ui'
import { toast } from 'react-hot-toast'
import api from '@/lib/api'

export default function OnboardingPage() {
  const { user, tenantId, loading } = useAuth()
  const [companyName, setCompanyName] = useState('')
  const [isProvisioning, setIsProvisioning] = useState(false)
  const router = useRouter()

  // Redirect if already provisioned
  useEffect(() => {
    if (!loading && user && tenantId) {
      router.push('/')
    }
  }, [user, tenantId, loading, router])

  const handleOnboarding = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsProvisioning(true)

    try {
      // Identity-to-Schema API Call
      await api.post('/onboarding/provision', {
        company_name: companyName
      })

      toast.success('Sovereign environment provisioned!')
      toast('Please log out and log back in to activate your schema.', {
        icon: '🔐',
        duration: 6000
      })
      
      // Force refresh of auth state
      router.push('/login')
    } catch (error: any) {
      toast.error(error.message || 'Onboarding failed')
    } finally {
      setIsProvisioning(false)
    }
  }

  if (loading) return null

  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4">
      <div className="max-w-md w-full">
        <div className="bg-slate-900 border border-slate-800 p-8 rounded-3xl shadow-2xl relative overflow-hidden">
          {/* Progress Indicator */}
          <div className="absolute top-0 left-0 w-full h-1 bg-slate-800">
            <div className="h-full bg-blue-600 w-1/3 transition-all duration-1000"></div>
          </div>

          <div className="flex flex-col items-center mb-8">
            <div className="w-12 h-12 bg-blue-600/20 rounded-xl flex items-center justify-center mb-4">
              <MaterialIcon name="auto_awesome" className="text-2xl text-blue-400" />
            </div>
            <h1 className="text-xl font-bold text-white tracking-tight text-center">Finalizing your Sovereign Portal</h1>
            <p className="text-slate-400 text-xs text-center mt-2 px-4 leading-relaxed">
              We need a few more details to provision your isolated Australian data environment.
            </p>
          </div>

          <form onSubmit={handleOnboarding} className="space-y-6">
            <div className="space-y-2">
              <label className="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1">Company Name</label>
              <Input 
                placeholder="e.g. Acme Corp AU" 
                value={companyName}
                onChange={(e) => setCompanyName(e.target.value)}
                className="bg-slate-800/50 border-slate-700 text-slate-100 placeholder:text-slate-600 focus-visible:ring-blue-500 rounded-xl h-12 text-sm"
                required
              />
            </div>

            <div className="bg-blue-600/5 border border-blue-600/10 p-4 rounded-xl">
              <div className="flex gap-3">
                <MaterialIcon name="info" className="text-blue-400 text-lg flex-shrink-0" />
                <p className="text-[11px] text-blue-200/70 leading-relaxed">
                  Upon completion, your dedicated <span className="text-blue-400 font-bold">PostgreSQL schema</span> will be initialized in the Sydney region.
                </p>
              </div>
            </div>

            <Button 
                type="submit" 
                className="w-full bg-blue-600 hover:bg-blue-500 text-white font-bold h-12 rounded-xl transition-all shadow-lg shadow-blue-600/20"
                disabled={isProvisioning}
            >
              {isProvisioning ? (
                <span className="flex items-center gap-2">
                  <MaterialIcon name="sync" className="animate-spin text-lg" /> Provisioning...
                </span>
              ) : 'Complete Infrastructure Setup'}
            </Button>
          </form>

          <button 
            onClick={() => router.push('/login')}
            className="w-full mt-6 text-[10px] font-bold text-slate-500 hover:text-slate-300 uppercase tracking-widest transition-colors"
          >
            Cancel and Log Out
          </button>
        </div>
      </div>
    </div>
  )
}
