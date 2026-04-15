export default function RequirementCard() {
  return (
    <div className="requirement-card">
      <div className="requirement-card-header">
        <div className="requirement-title">
          <div className="fingerprint-icon">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 1a9 9 0 0 0-9 9v7c0 1.66.4 3.22 1.1 4.6"/>
              <path d="M12 1a9 9 0 0 1 9 9v7c0 1.66-.4 3.22-1.1 4.6"/>
              <path d="M9 22h6"/>
              <path d="M12 11v.01"/>
              <path d="M12 16v.01"/>
              <path d="M8 11v.01"/>
              <path d="M16 11v.01"/>
              <path d="M8 16v.01"/>
              <path d="M16 16v.01"/>
            </svg>
          </div>
          <span>Draft: Requirement #1042</span>
        </div>
        <span className="status-badge">IN PROGRESS</span>
      </div>
      <div className="requirement-card-body">
        <div className="requirement-meta">
          <div>
            <div className="meta-label">ACTOR</div>
            <div className="meta-value">End-User (Mobile)</div>
          </div>
          <div>
            <div className="meta-label">COMPLEXITY</div>
            <div className="complexity-bar">
              <div className="complexity-fill" />
            </div>
          </div>
        </div>
        <blockquote className="requirement-quote">
          "As a user, I want to use FaceID or TouchID so that I can log in securely without typing my password every time."
        </blockquote>
      </div>
    </div>
  )
}
