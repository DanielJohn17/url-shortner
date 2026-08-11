import { Header } from '@/components/Header'
import { Hero } from '@/components/Hero'
import { HowItWorks } from '@/components/HowItWorks'
import { Footer } from '@/components/Footer'

export default function App() {
  return (
    <div className="flex min-h-screen flex-col bg-background">
      <Header />
      <main className="flex-1 pt-20">
        <Hero />
        <HowItWorks />
      </main>
      <Footer />
    </div>
  )
}
