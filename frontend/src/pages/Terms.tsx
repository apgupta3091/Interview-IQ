import { Link } from 'react-router-dom'

export default function Terms() {
  return (
    <div className="max-w-3xl mx-auto px-4 py-12 prose prose-neutral dark:prose-invert">
      <h1>Terms of Service</h1>
      <p className="text-sm text-muted-foreground">Last updated: March 2026</p>

      <h2>1. Acceptance</h2>
      <p>
        By creating an account or using Interview-IQ, you agree to these Terms of Service. If you
        do not agree, do not use the service.
      </p>

      <h2>2. Description of Service</h2>
      <p>
        Interview-IQ is a coding interview preparation tracker. It allows you to log solved
        problems, track skill scores with time-based decay, and receive AI-generated practice
        recommendations. Some features require a paid subscription.
      </p>

      <h2>3. Acceptable Use</h2>
      <p>You agree not to:</p>
      <ul>
        <li>Use the service for any unlawful purpose.</li>
        <li>Attempt to reverse-engineer, scrape, or abuse the API.</li>
        <li>Create multiple accounts to circumvent usage limits.</li>
        <li>Share your account credentials with others.</li>
      </ul>

      <h2>4. Subscriptions and Billing</h2>
      <p>
        Paid plans are billed monthly or annually via Stripe. Subscriptions automatically renew
        unless cancelled. You may cancel at any time through the billing portal — your access
        continues until the end of the current billing period, after which your account reverts to
        the free tier. No refunds are issued for partial billing periods.
      </p>

      <h2>5. No Warranty</h2>
      <p>
        Interview-IQ is provided "as is" without warranty of any kind. We do not guarantee that
        the service will be uninterrupted, error-free, or that it will improve your interview
        performance. Skill scores and AI recommendations are for informational purposes only.
      </p>

      <h2>6. Limitation of Liability</h2>
      <p>
        To the maximum extent permitted by law, Interview-IQ and its operators shall not be liable
        for any indirect, incidental, special, or consequential damages arising from your use of the
        service.
      </p>

      <h2>7. Termination</h2>
      <p>
        We reserve the right to suspend or terminate accounts that violate these terms. You may
        delete your account at any time by contacting{' '}
        <a href="mailto:support@interview-iq.com">support@interview-iq.com</a>.
      </p>

      <h2>8. Changes to Terms</h2>
      <p>
        We may update these terms from time to time. Continued use of the service after changes
        constitutes acceptance of the new terms. We will notify users of material changes via email.
      </p>

      <h2>9. Contact</h2>
      <p>
        Questions? Email{' '}
        <a href="mailto:support@interview-iq.com">support@interview-iq.com</a>.
      </p>

      <div className="mt-8 text-sm text-muted-foreground">
        <Link to="/" className="underline">Back to app</Link>
        {' · '}
        <Link to="/privacy" className="underline">Privacy Policy</Link>
      </div>
    </div>
  )
}
