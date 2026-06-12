import { Message } from '../types/comms';

// --- Vault KV secrets engine (DEMO ONLY) ----------------------------------
// The Anthropic API key is stored in Vault (see the Integrations → Anthropic
// modal). Read it from there at call time so the key lives in the secrets
// engine rather than the browser. Demo uses the dev root token directly.
const VAULT_ADDR = (typeof window !== 'undefined' && (window as any).GOTAK_CONFIG?.vaultUrl) || 'http://127.0.0.1:8200';
const VAULT_TOKEN = 'root';
const ANTHROPIC_KV_PATH = 'secret/data/gotak/anthropic';

async function readAnthropicKeyFromVault(): Promise<string> {
  try {
    const res = await fetch(`${VAULT_ADDR}/v1/${ANTHROPIC_KV_PATH}`, {
      headers: { 'X-Vault-Token': VAULT_TOKEN },
    });
    if (!res.ok) return '';
    const json = await res.json();
    return json?.data?.data?.api_key || '';
  } catch {
    return '';
  }
}

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
    this.model = config.model || 'claude-sonnet-4-6';
    this.maxTokens = config.maxTokens || 2048;
  }

  // Resolve the key: Vault KV is the source of truth, then any locally-provided
  // fallbacks (config / session cache / env).
  private async resolveApiKey(): Promise<string> {
    const fromVault = await readAnthropicKeyFromVault();
    if (fromVault) {
      this.apiKey = fromVault;
      return fromVault;
    }
    return this.apiKey;
  }

  async sendMessage(userMessage: string, conversationHistory: Message[] = []): Promise<AIResponse> {
    const apiKey = await this.resolveApiKey();
    if (!apiKey) {
      throw new Error('Anthropic API key not configured (store it in Vault via Integrations → Anthropic)');
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
          'x-api-key': apiKey,
          'anthropic-version': '2023-06-01',
          // Required for calling the Anthropic API directly from a browser.
          'anthropic-dangerous-direct-browser-access': 'true'
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

  // Check if service is configured (sync — only sees a locally-cached key).
  isConfigured(): boolean {
    return !!this.apiKey;
  }

  // Vault-aware configuration check: resolves the key from Vault KV (and caches
  // it) so the UI knows the officer is configured even when the key only lives
  // in Vault, not localStorage.
  async ensureConfigured(): Promise<boolean> {
    const key = await this.resolveApiKey();
    return !!key;
  }
}

// Export singleton instance
export const aiService = new AIService({
  apiKey: localStorage.getItem('anthropic_api_key') || import.meta.env.VITE_ANTHROPIC_API_KEY || '',
  model: import.meta.env.VITE_AI_MODEL || 'claude-sonnet-4-6',
  maxTokens: parseInt(import.meta.env.VITE_AI_MAX_TOKENS || '2048')
});

export default aiService;