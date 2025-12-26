// Type definitions matching backend models

export type UserRole = 'customer' | 'agent' | 'admin';

export interface User {
  id: string;
  tenant_id: string;
  email: string;
  role: UserRole;
  created_at: string;
  updated_at: string;
}

export interface Conversation {
  id: string;
  tenant_id: string;
  customer_id?: string;
  customer_email?: string;
  product_id?: string;
  status: 'active' | 'closed' | 'archived';
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: string;
  conversation_id: string;
  sender: 'customer' | 'agent';
  content: string;
  channel: string;
  language: string;
  timestamp: string;
  created_at: string;
}

export interface ConversationMetadata {
  id: string;
  conversation_id: string;
  intent: string;
  intent_score: number;
  sentiment: 'positive' | 'neutral' | 'negative';
  sentiment_score: number;
  emotions: string[];
  objections: string[];
  updated_at: string;
}

export interface ConversationWithMetadata extends Conversation {
  messages?: Message[];
  metadata?: ConversationMetadata;
}

export interface Rule {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  type: 'block' | 'correct' | 'flag';
  pattern: string;
  action: 'block' | 'auto_correct' | 'flag';
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CustomerMemory {
  id: string;
  tenant_id: string;
  customer_id: string;
  preferred_language: string;
  pricing_sensitivity: 'high' | 'medium' | 'low';
  product_interests: string[];
  past_objections: string[];
  created_at: string;
  updated_at: string;
}

export interface AISuggestion {
  reply: string;
  confidence: number;
  reasoning?: string;
  context?: string;
  product_match?: boolean;
  product_recommendations?: string[];
}

export interface SuggestionsResponse {
  suggestions: AISuggestion[];
  contextUsed?: boolean;
  context_used?: boolean; // Backend uses snake_case
  timing?: {
    recommended_delay_seconds: number;
    reason: string;
  };
  pricing?: {
    recommended_discount_percent?: number;
    reason: string;
  };
}

export interface LeadContext {
  source: string;
  product_interest?: string;
  customer_type: string;
  channel: string;
}

export interface AIInsights {
  intent: string;
  primary_objection?: string;
  sentiment_trend: string;
  confidence: number;
}

export interface EngagementMetrics {
  last_message_time: string;
  response_delay_risk: string;
  silence_detected: boolean;
}

export interface PrioritizedLead {
  conversation_id: string;
  customer_email?: string;
  score?: number;
  priority_score?: number;
  win_probability: number;
  urgency?: 'high' | 'medium' | 'low';
  urgency_score?: number;
  deal_value?: number;
  reasons?: string[];
  lead_context?: LeadContext;
  ai_insights?: AIInsights;
  engagement?: EngagementMetrics;
  recommended_action?: string;
  lead_stage?: string;
  risk_flags?: string[];
}

export interface WinProbability {
  probability: number;
  factors: string[];
  trend: 'improving' | 'stable' | 'deteriorating';
}

export interface ChurnRisk {
  risk_level: 'high' | 'medium' | 'low';
  score: number;
  factors: string[];
}

export interface TrendAnalysis {
  sentiment_trend: 'improving' | 'stable' | 'deteriorating';
  emotion_trend: 'improving' | 'stable' | 'deteriorating';
  engagement_trend: 'improving' | 'stable' | 'deteriorating';
  data_points: Array<{
    timestamp: string;
    sentiment_score: number;
    emotion_scores: Record<string, number>;
    engagement_score: number;
  }>;
}

export interface CLVEstimate {
  estimated_clv: number;
  confidence: 'high' | 'medium' | 'low';
  factors: string[];
}

export interface SalesCyclePrediction {
  predicted_days: number;
  confidence: 'high' | 'medium' | 'low';
  factors: string[];
}

export interface QualityScore {
  overall_score: number;
  response_time_score: number;
  sentiment_improvement_score: number;
  engagement_score: number;
  factors: string[];
}

export interface DashboardMetrics {
  total_conversations: number;
  active_conversations: number;
  average_sentiment: number;
  win_rate: number;
  churn_rate: number;
  top_intents: Array<{ intent: string; count: number }>;
  top_objections: Array<{ objection: string; count: number }>;
}

export interface Product {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  category: string;
  price: number;
  price_currency: string;
  features: string[];
  limitations: string[];
  target_audience: string;
  common_questions: string[];
  created_at: string;
  updated_at: string;
}

export interface AutoReplyGlobalConfig {
  tenant_id: string;
  enabled: boolean;
  confidence_threshold: number;
  updated_at: string;
}

export interface AutoReplyConversationConfig {
  conversation_id: string;
  enabled: boolean;
  confidence_threshold?: number;
  updated_at: string;
}

export interface EffectiveAutoReplyConfig {
  enabled: boolean;
  confidence_threshold: number;
  source: 'global' | 'conversation';
}

