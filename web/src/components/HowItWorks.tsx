import { ClipboardPaste, Minimize, Share2 } from 'lucide-react'

const steps = [
  {
    title: 'Paste',
    description: 'Copy your long, unwieldy URL and paste it into the field above.',
    icon: ClipboardPaste,
    color: 'text-primary bg-primary/10',
  },
  {
    title: 'Shorten',
    description: 'Click shorten to generate a clean short link instantly.',
    icon: Minimize,
    color: 'text-secondary bg-secondary/10',
  },
  {
    title: 'Share',
    description: 'Share your new link anywhere and start tracking engagement.',
    icon: Share2,
    color: 'text-orange-600 bg-orange-500/10',
  },
]

export function HowItWorks() {
  return (
    <section className="w-full border-t border-border/60 bg-card">
      <div className="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:px-12 lg:py-24">
        <div className="mb-12 text-center sm:mb-14">
          <h2 className="text-2xl font-bold tracking-tight sm:text-3xl lg:text-4xl">How it works</h2>
          <p className="mx-auto mt-3 max-w-2xl text-muted-foreground">
            Start optimizing your links in three simple steps. No complex setup required.
          </p>
        </div>

        <div className="relative grid grid-cols-1 gap-12 md:grid-cols-3 md:gap-6 lg:gap-12">
          <div className="pointer-events-none absolute top-12 right-[16%] left-[16%] hidden h-0.5 bg-gradient-to-r from-transparent via-border to-transparent md:block" />

          {steps.map((step) => (
            <div key={step.title} className="group relative flex flex-col items-center text-center">
              <div className="relative z-10 flex h-24 w-24 items-center justify-center rounded-2xl bg-muted/70 shadow-sm transition-transform duration-500 group-hover:-translate-y-2">
                <step.icon className={`size-10 ${step.color}`} />
              </div>
              <h3 className="mt-6 text-xl font-semibold sm:text-2xl">{step.title}</h3>
              <p className="mt-2 max-w-xs text-sm leading-6 text-muted-foreground sm:text-base">
                {step.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
