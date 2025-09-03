package Discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
	"time"
)

var prefix = "!"

func Dc() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN not found in environment variables")
	}

	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	sess.AddHandler(helpCommand)
	sess.AddHandler(handleButtonInteraction)
	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsMessageContent
	err = sess.Open()
	if err != nil {
		panic(err)
	}
	defer sess.Close()
	fmt.Println("Bot is running")

	// Keep the bot running
	select {}
}

func helpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, prefix) {
		return
	}

	// Only respond to !help command
	if m.Content != "!help" {
		return
	}

	// Get bot user info from session state
	var botAvatarURL string
	if s.State.User != nil {
		botAvatarURL = s.State.User.AvatarURL("")
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🤖 AI Assistant Control Panel",
		Description: "**Welcome to the Professional AI Management System**\n\nSelect a category below to view detailed information about available commands and features.",
		Color:       0x5865F2, // Discord's blurple color
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: botAvatarURL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "📋 Available Categories",
				Value: "🔧 **Channel Management** - Activate/Deactivate AI\n" +
					"💾 **Data Management** - Memory and history controls\n" +
					"⚙️ **Configuration** - Settings and preferences\n" +
					"📊 **Status & Info** - System information and statistics\n" +
					"❓ **Support** - Help and troubleshooting",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("AI Control Panel • Requested by %s • %s", m.Author.Username, time.Now().Format("Jan 02, 2006")),
			IconURL: m.Author.AvatarURL(""),
		},
	}

	// Create interactive buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: "help_channel_mgmt",
					Label:    "🔧 Channel Management",
					Style:    discordgo.PrimaryButton,
				},
				discordgo.Button{
					CustomID: "help_data_mgmt",
					Label:    "💾 Data Management",
					Style:    discordgo.SecondaryButton,
				},
				discordgo.Button{
					CustomID: "help_config",
					Label:    "⚙️ Configuration",
					Style:    discordgo.SecondaryButton,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: "help_status",
					Label:    "📊 Status & Info",
					Style:    discordgo.SecondaryButton,
				},
				discordgo.Button{
					CustomID: "help_support",
					Label:    "❓ Support",
					Style:    discordgo.SecondaryButton,
				},
				discordgo.Button{
					CustomID: "help_home",
					Label:    "🏠 Main Menu",
					Style:    discordgo.SuccessButton,
				},
			},
		},
	}

	_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Embed:      embed,
		Components: components,
	})
	if err != nil {
		log.Printf("Error sending help embed: %v", err)
	}
}
func handleButtonInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID

	var embed *discordgo.MessageEmbed
	var components []discordgo.MessageComponent

	// Get bot avatar
	var botAvatarURL string
	if s.State.User != nil {
		botAvatarURL = s.State.User.AvatarURL("")
	}

	switch customID {
	case "help_channel_mgmt":
		embed = &discordgo.MessageEmbed{
			Title:       "🔧 Channel Management Commands",
			Description: "**Control AI presence in your Discord channels**\n\nManage where and how the AI assistant operates within your server.",
			Color:       0x5865F2,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: botAvatarURL},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "🟢 Activate AI",
					Value:  "```!activate```\n**Function:** Enables AI responses in the current channel\n**Permission:** Requires Manage Channels permission\n**Usage:** The AI will begin responding to messages and participating in conversations",
					Inline: false,
				},
				{
					Name:   "🔴 Deactivate AI",
					Value:  "```!deactivate```\n**Function:** Disables AI responses in the current channel\n**Permission:** Requires Manage Channels permission\n**Usage:** The AI will stop responding but will retain conversation history",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{Text: "Channel Management • Use buttons below to navigate"},
		}

	case "help_data_mgmt":
		embed = &discordgo.MessageEmbed{
			Title:       "💾 Data Management Commands",
			Description: "**Manage AI memory and conversation data**\n\nControl how the AI stores and processes conversation history.",
			Color:       0x5865F2,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: botAvatarURL},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "🔄 Reset Memory",
					Value:  "```!wack```\n**Function:** Clears AI conversation history for this channel\n**Permission:** Requires Manage Messages permission\n**Usage:** Provides a fresh start - AI won't remember previous conversations\n**Warning:** This action cannot be undone",
					Inline: false,
				},
				{
					Name:   "📈 Memory Status",
					Value:  "```!memory```\n**Function:** Shows current memory usage and conversation count\n**Permission:** Available to all users\n**Usage:** Displays how much data the AI has stored for this channel",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{Text: "Data Management • Use buttons below to navigate"},
		}

	case "help_config":
		embed = &discordgo.MessageEmbed{
			Title:       "⚙️ Configuration Commands",
			Description: "**Customize AI behavior and settings**\n\nAdjust how the AI responds and behaves in your server.",
			Color:       0x5865F2,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: botAvatarURL},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "🎯 Response Mode",
					Value:  "```!mode [casual|professional|creative]```\n**Function:** Sets AI personality and response style\n**Options:** casual, professional, creative\n**Default:** professional",
					Inline: false,
				},
				{
					Name:   "⏱️ Response Delay",
					Value:  "```!delay [0-30]```\n**Function:** Sets delay in seconds before AI responds\n**Range:** 0-30 seconds\n**Default:** 2 seconds",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{Text: "Configuration • Use buttons below to navigate"},
		}

	case "help_status":
		embed = &discordgo.MessageEmbed{
			Title:       "📊 Status & Information Commands",
			Description: "**Monitor AI performance and system status**\n\nGet detailed information about the AI assistant's current state.",
			Color:       0x5865F2,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: botAvatarURL},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "🔍 System Status",
					Value:  "```!status```\n**Function:** Shows AI system health and performance metrics\n**Information:** Uptime, memory usage, response times\n**Refresh:** Updates every 30 seconds",
					Inline: false,
				},
				{
					Name:   "📋 Channel Info",
					Value:  "```!info```\n**Function:** Displays AI configuration for current channel\n**Shows:** Active status, memory usage, settings, last activity\n**Access:** Available to all users",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{Text: "Status & Information • Use buttons below to navigate"},
		}

	case "help_support":
		embed = &discordgo.MessageEmbed{
			Title:       "❓ Support & Troubleshooting",
			Description: "**Get help and resolve issues**\n\nFind solutions to common problems and get additional support.",
			Color:       0x5865F2,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: botAvatarURL},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "🆘 Emergency Commands",
					Value:  "```!emergency stop``` - Immediately halt AI in all channels\n```!emergency reset``` - Full system reset (admin only)\n```!emergency logs``` - Generate diagnostic report",
					Inline: false,
				},
				{
					Name:   "📞 Contact Support",
					Value:  "• **Server Admin:** Contact your Discord server administrator\n• **Technical Issues:** Use `!report [issue]` command\n• **Feature Requests:** Use `!suggest [idea]` command",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{Text: "Support & Troubleshooting • Use buttons below to navigate"},
		}

	default: // help_home or fallback
		embed = &discordgo.MessageEmbed{
			Title:       "🤖 AI Assistant Control Panel",
			Description: "**Welcome to the Professional AI Management System**\n\nSelect a category below to view detailed information about available commands and features.",
			Color:       0x5865F2,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: botAvatarURL},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "📋 Available Categories",
					Value: "🔧 **Channel Management** - Activate/Deactivate AI\n" +
						"💾 **Data Management** - Memory and history controls\n" +
						"⚙️ **Configuration** - Settings and preferences\n" +
						"📊 **Status & Info** - System information and statistics\n" +
						"❓ **Support** - Help and troubleshooting",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("AI Control Panel • Requested by %s • %s",
					i.Member.User.Username, time.Now().Format("Jan 02, 2006")),
				IconURL: i.Member.User.AvatarURL(""),
			},
		}
	}

	// Navigation buttons (same for all pages)
	components = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: "help_channel_mgmt",
					Label:    "🔧 Channel Management",
					Style:    discordgo.PrimaryButton,
				},
				discordgo.Button{
					CustomID: "help_data_mgmt",
					Label:    "💾 Data Management",
					Style:    discordgo.SecondaryButton,
				},
				discordgo.Button{
					CustomID: "help_config",
					Label:    "⚙️ Configuration",
					Style:    discordgo.SecondaryButton,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: "help_status",
					Label:    "📊 Status & Info",
					Style:    discordgo.SecondaryButton,
				},
				discordgo.Button{
					CustomID: "help_support",
					Label:    "❓ Support",
					Style:    discordgo.SecondaryButton,
				},
				discordgo.Button{
					CustomID: "help_home",
					Label:    "🏠 Main Menu",
					Style:    discordgo.SuccessButton,
				},
			},
		},
	}

	// Update the message
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
	}
}
