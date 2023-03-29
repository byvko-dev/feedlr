package router

import (
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Context struct {
	session *discordgo.Session
	data    *discordgo.InteractionCreate
}

// Returns true the handler was invoked by a slash command
func (c Context) IsCommand() bool {
	return c.data.Type == discordgo.InteractionApplicationCommand
}

// Returns true if the handler was invoked by a modal
func (c Context) IsModal() bool {
	return c.data.Type == discordgo.InteractionModalSubmit
}

// Returns the modal data
func (c Context) ModalData() (discordgo.ModalSubmitInteractionData, error) {
	if !c.IsModal() {
		return discordgo.ModalSubmitInteractionData{}, fmt.Errorf("not a modal")
	}
	return c.data.ModalSubmitData(), nil
}

// Get the guild where the command was invoked
func (c Context) Guild() (*discordgo.Guild, error) {
	return c.session.Guild(c.data.GuildID)
}

// Get the channel where the command was invoked
func (c Context) Channel() (*discordgo.Channel, error) {
	return c.session.Channel(c.data.ChannelID)
}

// Send a message to the channel where the command was invoked
func (c Context) Reply(content string) error {
	return c.session.InteractionRespond(c.data.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

// Send an embed to the channel where the command was invoked
func (c Context) ReplyEmbeds(embeds ...*discordgo.MessageEmbed) error {
	return c.session.InteractionRespond(c.data.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
		},
	})
}

// Send a modal to the channel where the command was invoked
func (c Context) ReplyModal(id string, title string, content []discordgo.MessageComponent) error {
	data := discordgo.InteractionResponseData{
		Title:      title,
		CustomID:   fmt.Sprintf("%v_%v", id, c.data.Member.User.ID),
		Components: content,
	}

	t, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(t))

	return c.session.InteractionRespond(c.data.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &data,
	})
}

// Create a webhook in the channel where the command was invoked
func (c Context) CreateWebhook(name string, avatar string) (*discordgo.Webhook, error) {
	return c.session.WebhookCreate(c.data.ChannelID, name, avatar)
}

// Delete a webhook by ID
func (c Context) DeleteWebhook(id string) error {
	return c.session.WebhookDelete(id)
}
