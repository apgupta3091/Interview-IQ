import { Link } from 'react-router-dom'

export default function PrivacyPolicy() {
  return (
    <div className="max-w-3xl mx-auto px-4 py-12 prose prose-neutral dark:prose-invert">
      <h1>Privacy Policy</h1>
      <p className="text-sm text-muted-foreground">Last updated: March 2026</p>

      <h2>1. What We Collect</h2>
      <p>
        We collect the following information when you use Interview-IQ:
      </p>
      <ul>
        <li>
          <strong>Account information</strong> — your email address, collected and managed by{' '}
          <a href="https://clerk.com" target="_blank" rel="noreferrer">Clerk</a>. We do not store
          your password directly.
        </li>
        <li>
          <strong>Problem logs</strong> — the coding problems you record, including problem name,
          category, difficulty, number of attempts, time taken, and whether you looked at the
          solution.
        </li>
        <li>
          <strong>Usage data</strong> — anonymous product analytics via PostHog (page views, feature
          interactions) to help us improve the product.
        </li>
        <li>
          <strong>Error data</strong> — crash reports and error traces via Sentry to help us fix
          bugs.
        </li>
      </ul>

      <h2>2. How We Use Your Data</h2>
      <ul>
        <li>To display your problem history and skill scores on your dashboard.</li>
        <li>To generate AI-powered practice recommendations.</li>
        <li>To process subscription payments via Stripe.</li>
        <li>To diagnose and fix bugs.</li>
        <li>To understand how users interact with the product (aggregate, anonymous).</li>
      </ul>

      <h2>3. Data Storage</h2>
      <p>
        Your problem logs are stored in a PostgreSQL database hosted on Railway. Backups are taken
        daily. Authentication data is stored and managed by Clerk. Payment information is handled
        entirely by Stripe — we never see or store your card details.
      </p>

      <h2>4. Data Sharing</h2>
      <p>
        We do not sell your data. We share data only with the service providers listed above (Clerk,
        Stripe, Railway, PostHog, Sentry, OpenAI) solely to operate the service.
      </p>

      <h2>5. Data Deletion</h2>
      <p>
        You may request deletion of your account and all associated data by emailing{' '}
        <a href="mailto:support@interview-iq.com">support@interview-iq.com</a>. We will process
        deletion requests within 30 days.
      </p>

      <h2>6. Cookies</h2>
      <p>
        We use cookies and local storage only for authentication (Clerk session token) and anonymous
        analytics. No third-party advertising cookies are used.
      </p>

      <h2>7. Contact</h2>
      <p>
        Questions about this policy? Email{' '}
        <a href="mailto:support@interview-iq.com">support@interview-iq.com</a>.
      </p>

      <div className="mt-8 text-sm text-muted-foreground">
        <Link to="/" className="underline">Back to app</Link>
        {' · '}
        <Link to="/terms" className="underline">Terms of Service</Link>
      </div>
    </div>
  )
}
