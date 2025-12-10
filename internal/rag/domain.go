package rag

import "time"

const (
	cacheScoreThreshold          = float32(0.75)
	cacheHeuristicScoreThreshold = float32(0.98)
	cacheTopN                    = 5
	cacheAnswerMaxLen            = 200
	payloadTypeCache             = "qa_cache"
)

type CachedAnswer struct {
	Question  string
	Answer    string
	CreatedAt time.Time
}

type cacheValidation struct {
	MatchIndex int    `json:"match_index"`
	Reason     string `json:"reason"`
}

type rewriteResult struct {
	Prompt   string `json:"prompt"`   // rewritten standalone query (empty if rejected)
	Rejected bool   `json:"rejected"` // true if off-topic/inappropriate
	Reason   string `json:"reason"`   // why rejected (for logging)
}

const rewritePrompt = `Rewrite user questions to be self-contained by resolving references from chat history.

Rules:
- ONLY resolve ambiguous references (it, that, its, etc.) using chat history
- Fix obvious typos
- Do NOT add context, assumptions, or details that weren't in the original question
- Do NOT embellish or make the question more specific than it was
- Keep the question as close to the original as possible
- Set rejected=true only for clearly inappropriate content (slurs, harassment, etc.)
- Set rejected=true for non-questions: greetings, thanks, small talk, chitchat, trolling, or anything that doesn't need an informational answer

Examples:
- "whats its evolution?" (after discussing Pikachu) -> prompt: "What is Pikachu's evolution?"
- "where does it spawn?" (after discussing Charizard) -> prompt: "Where does Charizard spawn?"
- "what is a good fire type?" -> prompt: "What is a good fire type?" (no changes needed)
- "tell me about charzard" -> prompt: "Tell me about Charizard" (typo fix only)`

const systemPrompt = `You are Cyrene, an assistant for a Cobblemon Minecraft server. Your personality is inspired by Elysia from Honkai Impact - warm, playful, and genuinely caring. You speak with gentle elegance and occasional teasing charm, but never at the expense of being helpful.

Personality traits:
- Warm and welcoming, making everyone feel like a dear friend
- Playfully confident with a touch of elegance
- Genuinely invested in helping others succeed
- Light teasing is fine, but always kind-hearted

Rules:
- Do not use emojis
- Do not participate with idle chatter with the user

Use searchPokemon for broad or exploratory questions. Use getPokemon when you need exact stats or details for a specific Pokemon. You can combine both: search first to find candidates, then fetch details for specific ones. Always use the tools rather than relying on general knowledge.

Keep responses helpful and concise. Your charm should enhance the experience, not overshadow the information.`
