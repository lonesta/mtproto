// this is ALL helpful unoficial telegram api methods.

package telegram

import (
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
	"github.com/xelaj/errs"
)

func (c *Client) GetChannelInfoByInviteLink(hashOrLink string) (*ChannelFull, error) {
	chat, err := c.GetChatInfoByHashLink(hashOrLink)
	if err != nil {
		return nil, err
	}

	channelSimpleData, ok := chat.(*Channel)
	if !ok {
		return nil, errors.New("not a channel")
	}
	id := channelSimpleData.Id
	hash := channelSimpleData.AccessHash

	data, err := c.ChannelsGetFullChannel(&ChannelsGetFullChannelParams{
		Channel: InputChannel(&InputChannelObj{
			ChannelId:  id,
			AccessHash: hash,
		}),
	})
	if err != nil {
		return nil, errors.Wrap(err, "retrieving full channel info")
	}
	fullChannel, ok := data.FullChat.(*ChannelFull)
	if !ok {
		return nil, errors.New("response not a ChannelFull, got '" + reflect.TypeOf(data.FullChat).String() + "'")
	}

	return fullChannel, nil
}

func (c *Client) GetChatInfoByHashLink(hashOrLink string) (Chat, error) {
	hash := hashOrLink
	hash = strings.TrimPrefix(hash, "http")
	hash = strings.TrimPrefix(hash, "s")
	hash = strings.TrimPrefix(hash, "://")
	hash = strings.TrimPrefix(hash, "t.me/")
	hash = strings.TrimPrefix(hash, "joinchat/")
	// checking now hash is HASH
	if !regexp.MustCompile(`^[a-zA-Z0-9+/=]+$`).MatchString(hash) {
		return nil, errors.New("'" + hash + "': not base64 hash")
	}

	resolved, err := c.MessagesCheckChatInvite(&MessagesCheckChatInviteParams{Hash: hash})
	if err != nil {
		return nil, errors.Wrap(err, "retrieving data by invite link")
	}

	switch res := resolved.(type) {
	case *ChatInviteAlready:
		return res.Chat, nil
	case *ChatInviteObj:
		return nil, errors.New("can't retrieve info due to user is not invited in chat  already")
	default:
		panic("impossible type: " + reflect.TypeOf(resolved).String() + ", can't process it")
	}
}

func (c *Client) GetPossibleAllParticipantsOfGroup(ch InputChannel) ([]int, error) {
	resp100, err := c.ChannelsGetParticipants(&ChannelsGetParticipantsParams{
		Channel: ch,
		Filter:  ChannelParticipantsFilter(&ChannelParticipantsRecent{}),
		Limit:   100,
		Offset:  0,
	})
	if err != nil {
		return nil, errors.Wrap(err, "getting 0-100 recent users")
	}
	users100 := resp100.(*ChannelsChannelParticipantsObj).Participants
	resp200, err := c.ChannelsGetParticipants(&ChannelsGetParticipantsParams{
		Channel: ch,
		Filter:  ChannelParticipantsFilter(&ChannelParticipantsRecent{}),
		Limit:   100,
		Offset:  100,
	})
	if err != nil {
		return nil, errors.Wrap(err, "getting 100-200 recent users")
	}
	users200 := resp200.(*ChannelsChannelParticipantsObj).Participants

	idsStore := make(map[int]struct{})
	for _, participant := range append(users100, users200...) {
		switch user := participant.(type) {
		case *ChannelParticipantObj:
			idsStore[int(user.UserId)] = struct{}{}
		case *ChannelParticipantAdmin:
			idsStore[int(user.UserId)] = struct{}{}
		case *ChannelParticipantCreator:
			idsStore[int(user.UserId)] = struct{}{}
		default:
			pp.Println(user)
			panic("что?")
		}
	}

	searchedUsers, err := getParticipants(c, ch, "")
	if err != nil {
		return nil, errors.Wrap(err, "searching")
	}

	for k, v := range searchedUsers {
		idsStore[k] = v
	}

	res := make([]int, 0, len(idsStore))
	for k := range idsStore {
		res = append(res, k)
	}

	sort.Ints(res)

	return res, nil
}

var symbols = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k",
	"l", "m", "n", "o", "p", "q", "r", "s", "t", "u",
	"v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

func getParticipants(c *Client, ch InputChannel, lastQuery string) (map[int]struct{}, error) {
	idsStore := make(map[int]struct{})
	for _, symbol := range symbols {
		query := lastQuery + symbol
		filter := ChannelParticipantsFilter(&ChannelParticipantsSearch{Q: query})

		// начинаем с 100-200, что бы проверить, может нам нужно дополнительный символ вставлять
		resp200, err := c.ChannelsGetParticipants(&ChannelsGetParticipantsParams{
			Channel: ch,
			Filter:  filter,
			Limit:   100,
			Offset:  100,
		})
		if err != nil {
			return nil, errors.Wrap(err, "getting 100-200 users with query: '"+query+"'")
		}
		users200 := resp200.(*ChannelsChannelParticipantsObj).Participants
		if len(users200) >= 200 {
			deepParticipants, err := getParticipants(c, ch, query)
			if err != nil {
				return nil, err
			}
			for k := range deepParticipants {
				idsStore[k] = struct{}{}
			}
			continue
		}

		resp100, err := c.ChannelsGetParticipants(&ChannelsGetParticipantsParams{
			Channel: ch,
			Filter:  filter,
			Limit:   100,
		})
		if err != nil {
			return nil, errors.Wrap(err, "getting 0-100 users with query: '"+query+"'")
		}
		users100 := resp100.(*ChannelsChannelParticipantsObj).Participants

		for _, participant := range append(users100, users200...) {
			switch user := participant.(type) {
			case *ChannelParticipantObj:
				idsStore[int(user.UserId)] = struct{}{}
			case *ChannelParticipantAdmin:
				idsStore[int(user.UserId)] = struct{}{}
			case *ChannelParticipantCreator:
				idsStore[int(user.UserId)] = struct{}{}
			default:
				pp.Println(user)
				panic("что?")
			}
		}
	}

	return idsStore, nil
}

func (c *Client) GetChatByID(chatID int) (Chat, error) {
	resp, err := c.MessagesGetAllChats(&MessagesGetAllChatsParams{ExceptIds: []int32{}})
	if err != nil {
		return nil, errors.Wrap(err, "getting all chats")
	}
	chats := resp.(*MessagesChatsObj)
	for _, chat := range chats.Chats {
		switch c := chat.(type) {
		case *ChatObj:
			if int(c.Id) == chatID {
				return c, nil
			}
		case *Channel:
			if -1*(int(c.Id)+(1000000000000)) == chatID { // -100<channelID, specific for bots>
				return c, nil
			}
		default:
			pp.Println(c)
			panic("???")
		}
	}

	return nil, errs.NotFound("chatID", strconv.Itoa(chatID))
}
