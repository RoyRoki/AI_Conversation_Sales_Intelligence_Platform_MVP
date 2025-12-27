import axios, { AxiosInstance } from 'axios';
import type {
  Conversation,
  ConversationWithMetadata,
  Message,
  Rule,
  CustomerMemory,
  SuggestionsResponse,
  PrioritizedLead,
  WinProbability,
  ChurnRisk,
  TrendAnalysis,
  CLVEstimate,
  SalesCyclePrediction,
  QualityScore,
  DashboardMetrics,
  Product,
  AutoReplyGlobalConfig,
  AutoReplyConversationConfig,
  EffectiveAutoReplyConfig,
} from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request interceptor to include auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = this.getToken();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Add response interceptor to handle token refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        if (error.response?.status === 401) {
          // Token expired or invalid
          this.clearToken();
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
          }
        }
        return Promise.reject(error);
      }
    );
  }

  private getToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('token');
  }

  private clearToken(): void {
    if (typeof window === 'undefined') return;
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  }

  // Authentication
  async login(email: string, password: string, tenantId: string): Promise<{ token: string; user: any }> {
    // Note: This assumes a login endpoint exists. If not, we'll need to add it to the backend.
    // For now, this is a placeholder that would need backend implementation.
    const response = await this.client.post('/api/auth/login', {
      email,
      password,
      tenant_id: tenantId,
    });
    const { token, user } = response.data;
    if (typeof window !== 'undefined') {
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
    }
    return { token, user };
  }

  logout(): void {
    this.clearToken();
  }

  // Customer login (email-only)
  async customerLogin(email: string, tenantId: string): Promise<{ token: string; user: any }> {
    const response = await this.client.post('/api/auth/customer-login', {
      email,
      tenant_id: tenantId,
    });
    const { token, user } = response.data;
    if (typeof window !== 'undefined') {
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
    }
    return { token, user };
  }

  // Conversations
  async createConversation(tenantId: string, productId?: string): Promise<Conversation> {
    const response = await this.client.post('/api/conversations', {
      tenant_id: tenantId,
      product_id: productId,
    });
    return response.data.conversation;
  }

  async getConversation(id: string): Promise<ConversationWithMetadata> {
    const response = await this.client.get(`/api/conversations/${id}`);
    // Handle both response structures: {conversation: {...}, messages: [...]} or {conversation: {..., messages: [...]}}
    const data = response.data;
    if (data.conversation) {
      // If messages are at root level, merge them into conversation
      if (data.messages && !data.conversation.messages) {
        data.conversation.messages = data.messages;
      }
      return data.conversation;
    }
    // Fallback: assume the whole response is the conversation
    return data;
  }

  async listConversations(): Promise<Conversation[]> {
    const response = await this.client.get('/api/conversations');
    return response.data.conversations || [];
  }

  async sendMessage(
    conversationId: string,
    sender: 'customer' | 'agent',
    message: string,
    channel?: string
  ): Promise<{ message_id: string; conversation_id: string; status: string }> {
    const response = await this.client.post(`/api/conversations/${conversationId}/messages`, {
      sender,
      message,
      channel: channel || 'web',
      timestamp: new Date().toISOString(),
    });
    return response.data;
  }

  // Agent Assist
  async getSuggestions(conversationId: string, regenerate: boolean = false): Promise<SuggestionsResponse> {
    try {
      const params = regenerate ? { regenerate: 'true' } : {};
      const response = await this.client.post(`/api/conversations/${conversationId}/suggestions`, {}, { params });
      const backendResponse = response.data.suggestions || response.data;
      
      // Handle case where backend returns error in response
      if (backendResponse.error) {
        console.warn('Backend returned error in suggestions response:', backendResponse.error);
        return {
          suggestions: [],
          contextUsed: false,
        };
      }
      
      // Transform backend Suggestion format to frontend AISuggestion format
      // Backend uses: { text, confidence, reasoning, product_match, product_recommendations }
      // Frontend expects: { reply, confidence, reasoning, product_match, product_recommendations }
      const suggestionsArray = backendResponse.suggestions || [];
      const transformedSuggestions = suggestionsArray.map((sug: any) => ({
        reply: sug.text || sug.reply || '',
        confidence: sug.confidence || 0,
        reasoning: sug.reasoning,
        product_match: sug.product_match,
        product_recommendations: sug.product_recommendations || [],
      }));

      return {
        suggestions: transformedSuggestions,
        contextUsed: backendResponse.context_used || backendResponse.contextUsed || false,
      };
    } catch (error: any) {
      console.error('Failed to get suggestions:', error);
      // Log more details for debugging
      if (error.response) {
        console.error('Error response status:', error.response.status);
        console.error('Error response data:', error.response.data);
      } else if (error.request) {
        console.error('No response received:', error.request);
      }
      // Return empty suggestions on error to allow graceful degradation
      return {
        suggestions: [],
        contextUsed: false,
      };
    }
  }

  async getInsights(conversationId: string): Promise<SuggestionsResponse> {
    const response = await this.client.get(`/api/conversations/${conversationId}/insights`);
    return response.data.insights;
  }

  // Analytics
  async getLeads(conversationIds?: string[]): Promise<{ leads: PrioritizedLead[]; total: number }> {
    const params = conversationIds ? { conversation_ids: conversationIds.join(',') } : {};
    const response = await this.client.get('/api/analytics/leads', { params });
    return response.data;
  }

  async getWinProbability(conversationId: string): Promise<WinProbability> {
    const response = await this.client.get(`/api/analytics/conversations/${conversationId}/win-probability`);
    return response.data.win_probability;
  }

  async getChurnRisk(conversationId: string): Promise<ChurnRisk> {
    const response = await this.client.get(`/api/analytics/conversations/${conversationId}/churn-risk`);
    return response.data.churn_risk;
  }

  async getTrends(conversationId: string): Promise<TrendAnalysis> {
    const response = await this.client.get(`/api/analytics/conversations/${conversationId}/trends`);
    return response.data.trends;
  }

  async getCLV(conversationId: string): Promise<CLVEstimate> {
    const response = await this.client.get(`/api/analytics/conversations/${conversationId}/clv`);
    return response.data.clv;
  }

  async getSalesCycle(conversationId: string): Promise<SalesCyclePrediction> {
    const response = await this.client.get(`/api/analytics/conversations/${conversationId}/sales-cycle`);
    return response.data.sales_cycle;
  }

  async getQuality(conversationId: string): Promise<QualityScore> {
    const response = await this.client.get(`/api/analytics/conversations/${conversationId}/quality`);
    return response.data.quality;
  }

  async getDashboard(): Promise<DashboardMetrics> {
    const response = await this.client.get('/api/analytics/dashboard');
    return response.data.metrics;
  }

  // Rules (Admin only)
  async listRules(): Promise<Rule[]> {
    const response = await this.client.get('/api/rules');
    return response.data.rules || [];
  }

  async getRule(id: string): Promise<Rule> {
    const response = await this.client.get(`/api/rules/${id}`);
    return response.data.rule;
  }

  async createRule(rule: Omit<Rule, 'id' | 'created_at' | 'updated_at'>): Promise<Rule> {
    const response = await this.client.post('/api/rules', rule);
    return response.data.rule;
  }

  async updateRule(id: string, rule: Partial<Rule>): Promise<Rule> {
    const response = await this.client.put(`/api/rules/${id}`, rule);
    return response.data.rule;
  }

  async deleteRule(id: string): Promise<void> {
    await this.client.delete(`/api/rules/${id}`);
  }

  // Products
  async listProducts(tenantId?: string): Promise<Product[]> {
    const params = tenantId ? { tenant_id: tenantId } : {};
    const response = await this.client.get('/api/products', { params });
    return response.data.products || [];
  }

  async getProduct(id: string, tenantId?: string): Promise<Product> {
    const params = tenantId ? { tenant_id: tenantId } : {};
    const response = await this.client.get(`/api/products/${id}`, { params });
    return response.data.product;
  }

  async createProduct(product: Omit<Product, 'id' | 'created_at' | 'updated_at' | 'tenant_id'>): Promise<Product> {
    const response = await this.client.post('/api/products', product);
    return response.data.product;
  }

  async updateProduct(id: string, product: Partial<Product>): Promise<Product> {
    const response = await this.client.put(`/api/products/${id}`, product);
    return response.data.product;
  }

  async deleteProduct(id: string): Promise<void> {
    await this.client.delete(`/api/products/${id}`);
  }

  // Auto-Reply
  async getGlobalAutoReply(): Promise<AutoReplyGlobalConfig> {
    const response = await this.client.get('/api/autoreply/global');
    return response.data.config;
  }

  async updateGlobalAutoReply(config: { enabled: boolean; confidence_threshold: number }): Promise<AutoReplyGlobalConfig> {
    const response = await this.client.put('/api/autoreply/global', config);
    return response.data.config;
  }

  async getConversationAutoReply(conversationId: string): Promise<{ config?: AutoReplyConversationConfig; effective: EffectiveAutoReplyConfig }> {
    const response = await this.client.get(`/api/conversations/${conversationId}/autoreply`);
    const data = response.data;
    // Ensure effective config is properly formatted (handle both camelCase and snake_case)
    if (data.effective) {
      data.effective = {
        enabled: data.effective.enabled ?? data.effective.Enabled ?? false,
        confidence_threshold: data.effective.confidence_threshold ?? data.effective.ConfidenceThreshold ?? 0,
        source: (data.effective.source ?? data.effective.Source ?? 'conversation') as 'global' | 'conversation',
      };
    }
    return data;
  }

  async updateConversationAutoReply(conversationId: string, config: { enabled: boolean; confidence_threshold?: number }): Promise<{ config: AutoReplyConversationConfig; effective: EffectiveAutoReplyConfig }> {
    const response = await this.client.put(`/api/conversations/${conversationId}/autoreply`, config);
    const data = response.data;
    // Ensure effective config is properly formatted (handle both camelCase and snake_case)
    if (data.effective) {
      data.effective = {
        enabled: data.effective.enabled ?? data.effective.Enabled ?? false,
        confidence_threshold: data.effective.confidence_threshold ?? data.effective.ConfidenceThreshold ?? 0,
        source: (data.effective.source ?? data.effective.Source ?? 'conversation') as 'global' | 'conversation',
      };
    }
    return data;
  }

  async testAutoReply(conversationId: string): Promise<{ would_send: boolean; confidence: number; suggestion: string; reason: string }> {
    const response = await this.client.post(`/api/conversations/${conversationId}/autoreply/test`);
    return response.data;
  }

  // Customer Memory (Admin only)
  async listMemories(limit?: number, offset?: number): Promise<{ memories: CustomerMemory[]; total: number }> {
    const params: any = {};
    if (limit !== undefined) params.limit = limit;
    if (offset !== undefined) params.offset = offset;
    const response = await this.client.get('/api/memories', { params });
    return response.data;
  }

  async getMemory(id: string): Promise<CustomerMemory> {
    const response = await this.client.get(`/api/memories/${id}`);
    return response.data.memory;
  }

  async createMemory(memory: Omit<CustomerMemory, 'id' | 'tenant_id' | 'created_at' | 'updated_at'>): Promise<CustomerMemory> {
    const response = await this.client.post('/api/memories', memory);
    return response.data.memory;
  }

  async updateMemory(id: string, memory: Partial<Omit<CustomerMemory, 'id' | 'tenant_id' | 'customer_id' | 'created_at' | 'updated_at'>>): Promise<CustomerMemory> {
    const response = await this.client.put(`/api/memories/${id}`, memory);
    return response.data.memory;
  }

  async deleteMemory(id: string): Promise<void> {
    await this.client.delete(`/api/memories/${id}`);
  }
}

export const apiClient = new ApiClient();

