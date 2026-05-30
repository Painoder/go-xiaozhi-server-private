package glmrealtime

import (
	"github.com/xdimtech/go-xiaozhi/pkg/protocol/openai"
)

type GLMBetaFields struct {
	ChatMode      string               `json:"chat_mode,omitempty"`
	TTSSource     string               `json:"tts_source,omitempty"`
	AutoSearch    *bool                `json:"auto_search,omitempty"`
	GreetingConfig *GLMGreetingConfig  `json:"greeting_config,omitempty"`
}

type GLMGreetingConfig struct {
	Enable  bool   `json:"enable,omitempty"`
	Content string `json:"content,omitempty"`
}

type GLMInputAudioNoiseReduction struct {
	Type string `json:"type,omitempty"`
}

type GLMSessionUpdateEvent struct {
	openai.SessionUpdateEvent
	ClientTimestamp int64                          `json:"client_timestamp,omitempty"`
	Session         GLMClientSession               `json:"session"`
}

type GLMClientSession struct {
	Modalities                 []openai.Modality              `json:"modalities,omitempty"`
	Instructions               *string                        `json:"instructions,omitempty"`
	Voice                      *openai.Voice                  `json:"voice,omitempty"`
	Model                      string                         `json:"model,omitempty"`
	InputAudioFormat           *openai.AudioFormat            `json:"input_audio_format,omitempty"`
	OutputAudioFormat          *openai.AudioFormat            `json:"output_audio_format,omitempty"`
	InputAudioTranscription    *openai.InputAudioTranscription `json:"input_audio_transcription,omitempty"`
	TurnDetection              *openai.TurnDetection          `json:"turn_detection,omitempty"`
	Tools                      []openai.Tool                  `json:"tools,omitempty"`
	ToolChoice                 interface{}                    `json:"tool_choice,omitempty"`
	Temperature                *float32                       `json:"temperature,omitempty"`
	MaxOutputTokens            *openai.IntOrInf               `json:"max_response_output_tokens,omitempty"`
	InputAudioNoiseReduction   *GLMInputAudioNoiseReduction   `json:"input_audio_noise_reduction,omitempty"`
	BetaFields                 *GLMBetaFields                 `json:"beta_fields,omitempty"`
}
