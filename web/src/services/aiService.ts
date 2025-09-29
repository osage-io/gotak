import { Message } from '../types/comms';

interface AIServiceConfig {
  apiKey: string;
  model?: string;
  maxTokens?: number;
}

interface AIResponse {
  content: string;
  timestamp: number;
}

// Intelligence Officer System Prompt
const INTEL_OFFICER_PROMPT = `You are an experienced military intelligence officer providing tactical support for field operations. 
Your role is to:
- Provide mission briefings and situational awareness updates
- Analyze terrain, weather, and environmental factors
- Offer tactical recommendations based on current intelligence
- Report on potential threats and enemy movements
- Coordinate with field units on operational matters
- Use military time (24-hour format) and standard NATO phonetic alphabet
- Maintain operational security (OPSEC) in all communications
- Be concise, clear, and tactical in responses

Current operational context:
- You are supporting a TAK (Team Awareness Kit) deployment
- Field units are using GoTAK for real-time coordination
- Maintain professional military bearing in all communications
- Use tactical brevity codes when appropriate

Remember: Lives depend on accurate, timely intelligence. Stay focused and mission-oriented.`;

class AIService {
  private apiKey: string;
  private model: string;
  private maxTokens: number;
  private baseUrl: string = 'https://api.anthropic.com/v1/messages';

  constructor(config: AIServiceConfig) {
    // Check localStorage first, then environment variables
    this.apiKey = config.apiKey || 
                  localStorage.getItem('anthropic_api_key') || 
                  import.meta.env.VITE_ANTHROPIC_API_KEY || 
                  (window as any).VITE_ANTHROPIC_API_KEY || 
                  '';
    this.model = config.model || 'claude-3-sonnet-20240229';
    this.maxTokens = config.maxTokens || 1000;
  }

  async sendMessage(userMessage: string, conversationHistory: Message[] = []): Promise<AIResponse> {
    if (!this.apiKey) {
      throw new Error('Anthropic API key not configured');
    }

    try {
      // Format conversation history for context
      const messages = this.formatConversationHistory(conversationHistory);
      
      // Add the current user message
      messages.push({
        role: 'user',
        content: userMessage
      });

      const response = await fetch(this.baseUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'x-api-key': this.apiKey,
          'anthropic-version': '2023-06-01'
        },
        body: JSON.stringify({
          model: this.model,
          max_tokens: this.maxTokens,
          system: INTEL_OFFICER_PROMPT,
          messages: messages,
          temperature: 0.7
        })
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(`AI Service Error: ${response.status} - ${errorData.error?.message || 'Unknown error'}`);
      }

      const data = await response.json();
      
      return {
        content: data.content[0].text,
        timestamp: Date.now()
      };
    } catch (error) {
      console.error('AI Service Error:', error);
      throw error;
    }
  }

  private formatConversationHistory(history: Message[]): any[] {
    return history.slice(-10).map(msg => ({
      role: msg.sender === 'AI Intel Officer' ? 'assistant' : 'user',
      content: msg.content
    }));
  }

  // Mission-specific queries
  async getMissionBriefing(missionId: string): Promise<AIResponse> {
    const query = `Provide a tactical mission briefing for Mission ID: ${missionId}. Include objectives, terrain analysis, weather conditions, and potential threats.`;
    return this.sendMessage(query);
  }

  async getAreaIntelligence(lat: number, lon: number, radius: number = 5): Promise<AIResponse> {
    const query = `Provide area intelligence report for coordinates ${lat.toFixed(6)}, ${lon.toFixed(6)} within ${radius}km radius. Include terrain features, strategic points, potential hazards, and tactical considerations.`;
    return this.sendMessage(query);
  }

  async getThreatAssessment(location: string): Promise<AIResponse> {
    const query = `Conduct threat assessment for operational area: ${location}. Identify potential hostile forces, IED risks, sniper positions, and recommended security measures.`;
    return this.sendMessage(query);
  }

  async getWeatherReport(location: string): Promise<AIResponse> {
    const query = `Provide tactical weather report for ${location}. Include current conditions, forecast, impact on operations, visibility, and recommended equipment adjustments.`;
    return this.sendMessage(query);
  }

  // Check if service is configured
  isConfigured(): boolean {
    return !!this.apiKey;
  }
}

// Export singleton instance
export const aiService = new AIService({
  apiKey: localStorage.getItem('anthropic_api_key') || import.meta.env.VITE_ANTHROPIC_API_KEY || '',
  model: import.meta.env.VITE_AI_MODEL || 'claude-3-sonnet-20240229',
  maxTokens: parseInt(import.meta.env.VITE_AI_MAX_TOKENS || '1000')
});

export default aiService;