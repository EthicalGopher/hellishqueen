package Discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"hellish/AI"
	"hellish/Database"
	"log"
	"os"
	"strings"
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
	sess.AddHandler(handleChat)
	sess.AddHandler(helpCommand)
	sess.AddHandler(handleButtonInteraction)
	sess.AddHandler(activeCommand)
	sess.AddHandler(handleSystemMessage)
	sess.AddHandler(handleAPI)
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

// helpCommand updated to show only implemented commands.
func helpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Content != prefix+"help" {
		return
	}

	var botAvatarURL string
	if s.State.User != nil {
		botAvatarURL = s.State.User.AvatarURL("")
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Help Command",
		Description: "**Hellish Queen**\n\nSelect a category below to view the commands you can use to configure and interact with me.",
		Color:       0x5865F2, // Discord's blurple color
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: botAvatarURL,
		},
		Image: &discordgo.MessageEmbedImage{
			URL: "https://media.discordapp.net/attachments/1360608003010728219/1413557061144547510/hellish-ezgif.com-optimize.gif?ex=68bc5d19&is=68bb0b99&hm=2cd2bec26267d3b4fd0e7f5a334a1be3412c2472567b387f6cd5af623e018b95&=",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📋 Available Categories",
				Value:  "🔧 **Channel Management**\n⚙️ **Configuration**",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("AI Control Panel • Requested by %s", m.Author.Username),
			IconURL: m.Author.AvatarURL(""),
		},
	}

	// Define the dropdown menu options for implemented commands
	options := []discordgo.SelectMenuOption{
		{
			Label:       "Channel Management",
			Value:       "help_channel_mgmt",
			Description: "Commands to control where the AI is active.",
		},
		{
			Label:       "Configuration",
			Value:       "help_config",
			Description: "Commands to configure AI behavior and API keys.",
		},
	}

	selectMenu := discordgo.SelectMenu{
		CustomID:    "category_select",
		Placeholder: "View commands by their category",
		Options:     options,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{selectMenu},
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

	switch i.Type {

	case discordgo.InteractionMessageComponent:
		handleComponentInteraction(s, i)

	case discordgo.InteractionModalSubmit:
		handleModalSubmit(s, i)
	}
}

// handleComponentInteraction updated to handle the new help menu.
func handleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	customID := data.CustomID

	if customID == "add_api_key_button" {
		perms, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
		if err != nil {
			log.Printf("Error getting user permissions for %s: %v", i.Member.User.ID, err)
			return
		}
		if perms&discordgo.PermissionManageGuild == 0 {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You need the `Manage Server` permission to use this button.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "api_key_modal",
				Title:    "Add New API Key",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "api_key_input",
								Label:       "Gemini API Key",
								Style:       discordgo.TextInputShort,
								Placeholder: "Enter your key here. It will not be shown publicly.",
								Required:    true,
								MinLength:   39,
								MaxLength:   40,
							},
						},
					},
				},
			},
		})
		if err != nil {
			log.Printf("Error showing modal: %v", err)
		}
		return
	}

	if customID == "category_select" {
		var embed *discordgo.MessageEmbed
		var botAvatarURL string
		if s.State.User != nil {
			botAvatarURL = s.State.User.AvatarURL("")
		}

		selectedValue := data.Values[0]
		switch selectedValue {
		case "help_channel_mgmt":
			embed = &discordgo.MessageEmbed{
				Title:       "🔧 Channel Management Commands",
				Description: "Control where I am active on this server.",
				Color:       0x5865F2,
				Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: botAvatarURL},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "🟢 `!activate`",
						Value: "**Function:** Enables me to respond to messages in the current channel.\n**Permission:** `Manage Server`",
					},
				},
			}
		case "help_config":
			embed = &discordgo.MessageEmbed{
				Title:       "⚙️ Configuration Commands",
				Description: "Customize my behavior and manage API keys.",
				Color:       0x5865F2,
				Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: botAvatarURL},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "🔑 `!api <add|view|remove|clear>`",
						Value: "**Function:** Manages the Gemini API keys I use for this server.\n• `add`: Opens a secure pop-up to add a key.\n• `view`: Shows a count of registered keys.\n• `remove <key>`: Removes a specific key.\n• `clear`: Removes all keys.\n**Permission:** `Manage Server` for modifying commands.",
					},
					{
						Name:  "📝 `!system <set|view|clear>`",
						Value: "**Function:** Manages the custom instructions I use for this server.\n• `set <message>`: Sets the system message.\n• `view`: Shows the current message.\n• `clear`: Clears the message.\n**Permission:** `Manage Server` for modifying commands.",
					},
				},
			}
		}

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: i.Message.Components,
			},
		})
		if err != nil {
			log.Printf("Error responding to help menu interaction: %v", err)
		}
	}
}

func handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()

	// Ensure we're handling the correct modal
	if data.CustomID != "api_key_modal" {
		return
	}

	apiKey := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	err := Database.AddAPIKey(i.GuildID, apiKey)
	if err != nil {
		log.Printf("Error adding API key via modal for guild %s: %v", i.GuildID, err)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ An error occurred while saving the API key. It might already be in the list, or the database is unavailable.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ API key has been added successfully and securely.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Error sending modal confirmation: %v", err)
	}
}

func activeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if !strings.HasPrefix(m.Content, prefix) {
		return
	}
	if m.Content != "!activate" {
		return
	}
	channelID, err := Database.FindChannel(m.GuildID)
	if err != nil {
		log.Println(err)
	}
	if channelID == "" || channelID != m.ChannelID {
		err := Database.InsertChannel(m.GuildID, m.ChannelID)
		if err != nil {
			log.Fatal(err)
		}
		_, err = s.ChannelMessageSend(m.ChannelID, "AI is now active in this channel")
		if err != nil {
			return
		}
	} else {
		_, err = s.ChannelMessageSend(m.ChannelID, "AI is already active in this channel")
		if err != nil {
			return
		}
	}
}

func handleChat(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if strings.HasPrefix(m.Content, prefix) {
		return
	}

	channelId, err := Database.FindChannel(m.GuildID)
	if err != nil {
		return
	}
	if channelId != m.ChannelID {
		return
	}
	systemMessage, err := Database.ViewSystemMessage(m.GuildID)
	if err != nil {
		return
	}
	input :=
		`
		UserInput :
		` + m.Content +
			`
		SystemMessage :
		` + systemMessage + `
			user name : ` + m.Author.Username + `
		`
	res, err := AI.Response(m.GuildID, AI.GetBasePersona(), input)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
	}
	_, err = s.ChannelMessageSend(m.ChannelID, res)
	if err != nil {
		return
	}
}

func handleSystemMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// We only care about messages starting with "!system"
	if !strings.HasPrefix(m.Content, prefix+"system") {
		return
	}

	parts := strings.Fields(m.Content)
	// The command should be at least `!system <subcommand>`
	if len(parts) < 2 {
		// Send usage info if just `!system` is typed
		s.ChannelMessageSend(m.ChannelID, "Usage: `!system <set|view|clear> [message]`")
		return
	}

	subcommand := parts[1]

	switch subcommand {
	case "set":
		// Check for 'Manage Server' permission
		perms, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
		if err != nil {
			log.Printf("Error getting user permissions for %s: %v", m.Author.ID, err)
			s.ChannelMessageSend(m.ChannelID, "Could not verify your permissions. Please try again.")
			return
		}
		if perms&discordgo.PermissionManageGuild == 0 {
			s.ChannelMessageSend(m.ChannelID, "You need the `Manage Server` permission to set the system message.")
			return
		}

		if len(parts) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Please provide a message to set. Usage: `!system set <your system message>`")
			return
		}

		message := strings.Join(parts[2:], " ")
		err = Database.InsertSystemMessage(m.GuildID, message)
		if err != nil {
			log.Printf("Error setting system message for guild %s: %v", m.GuildID, err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while updating the system message.")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "✅ System message has been updated successfully.")

	case "view":
		message, err := Database.ViewSystemMessage(m.GuildID)
		if err != nil {
			log.Printf("Error viewing system message for guild %s: %v", m.GuildID, err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while retrieving the system message.")
			return
		}

		displayMessage := "No system message is currently set."
		if message != "" {
			displayMessage = message
		}
		s.ChannelMessageSend(m.ChannelID, displayMessage)
	}
}
func handleAPI(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// We only care about messages starting with "!api"
	if !strings.HasPrefix(m.Content, prefix+"api") {
		return
	}

	parts := strings.Fields(m.Content)
	// The command should be at least `!api <subcommand>`
	if len(parts) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: `!api <add|view|remove|clear> [key]`")
		return
	}

	subcommand := parts[1]

	// Check for 'Manage Server' permission for any command that modifies data
	isModifyingCommand := subcommand == "add" || subcommand == "remove" || subcommand == "clear"
	if isModifyingCommand {
		perms, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
		if err != nil {
			log.Printf("Error getting user permissions for %s: %v", m.Author.ID, err)
			s.ChannelMessageSend(m.ChannelID, "Could not verify your permissions. Please try again.")
			return
		}
		if perms&discordgo.PermissionManageGuild == 0 {
			s.ChannelMessageSend(m.ChannelID, "You need the `Manage Server` permission to modify API keys.")
			return
		}
	}

	switch subcommand {
	case "add":
		// The permission check is already handled above.
		// Now, we send a message with a button to trigger the modal.
		_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content: "Click the button below to add a new API key securely via a pop-up form.",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Add API Key",
							Style:    discordgo.PrimaryButton,
							CustomID: "add_api_key_button", // A unique ID for our button
						},
					},
				},
			},
		})
		if err != nil {
			log.Printf("Error sending API key button: %v", err)
		}
		// The rest of the logic is now handled by the interaction handler below.
		return

	case "view":
		// Note: This relies on the new ViewAPIKeys function suggested below.
		keys, err := Database.ViewAPIKeys(m.GuildID)
		if err != nil {
			log.Printf("Error viewing API keys for guild %s: %v", m.GuildID, err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while retrieving API keys.")
			return
		}

		var keyList strings.Builder
		if len(keys) == 0 {
			keyList.WriteString("No API keys are currently set for this server.")
		} else {
			for i, key := range keys {
				// Mask the key for security, showing only the first and last 4 characters
				maskedKey := key
				if len(key) > 8 {
					maskedKey = fmt.Sprintf("%s...%s", key[:4], key[len(key)-4:])
				}
				keyList.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, maskedKey))
			}
		}

		embed := &discordgo.MessageEmbed{
			Title:       "🔑 Registered API Keys",
			Description: "These keys are used by the AI for generating responses.",
			Color:       0x5865F2, // Discord Blurple
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Keys on this Server",
					Value: keyList.String(),
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    fmt.Sprintf("Requested by %s", m.Author.Username),
				IconURL: m.Author.AvatarURL(""),
			},
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	case "remove":
		if len(parts) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Please provide the full API key to remove. Usage: `!api remove <api_key_to_remove>`")
			return
		}
		apiKeyToRemove := parts[2]
		// Note: This relies on the new RemoveAPIKey function suggested below.
		err := Database.RemoveAPIKey(m.GuildID, apiKeyToRemove)
		if err != nil {
			log.Printf("Error removing API key for guild %s: %v", m.GuildID, err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Make sure you provided the exact key to remove.")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "✅ API key has been removed successfully.")

	case "clear":
		// Note: This relies on the new ClearAPIKeys function suggested below.
		err := Database.ClearAPIKeys(m.GuildID)
		if err != nil {
			log.Printf("Error clearing API keys for guild %s: %v", m.GuildID, err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while clearing API keys.")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "✅ All API keys for this server have been cleared.")

	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unknown subcommand `%s`. Use `!api <add|view|remove|clear>`.", subcommand))
	}
}
