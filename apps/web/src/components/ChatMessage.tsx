import RequirementCard from './RequirementCard'

interface ChatMessageProps {
  type: 'ai' | 'user'
  content: React.ReactNode
  showActions?: boolean
  showCard?: boolean
}

export default function ChatMessage({ type, content, showActions, showCard }: ChatMessageProps) {
  return (
    <div className={`chat-message ${type}`}>
      <div className="message-avatar">
        {type === 'ai' ? (
          <div className="ai-avatar">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 3l1.5 4.5h4.5l-3.75 2.75 1.5 4.5-3.75-2.75-3.75 2.75 1.5-4.5L6 7.5h4.5z"/>
            </svg>
          </div>
        ) : (
          <div className="user-avatar-small">JD</div>
        )}
      </div>
      <div className="message-content">
        <div className="message-bubble">{content}</div>
        {showActions && (
          <div className="message-actions">
            <button className="action-chip">User Story Mapping</button>
            <button className="action-chip">Technical Specification</button>
            <button className="action-chip">SLA Definition</button>
          </div>
        )}
        {showCard && <RequirementCard />}
      </div>
    </div>
  )
}
