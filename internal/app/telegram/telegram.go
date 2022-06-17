package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"
)

const (
	defaultTimeout = 100
)

type Telegram interface {
	Start(ctx context.Context) error
}

type telegram struct {
	httpClient *http.Client
	token      string
	timeout    uint
	chats      map[int64]chan string
	mu         sync.RWMutex
	userClient UserClient
	repository Repository
}

func New(token string, userClient UserClient, repository Repository) *telegram {
	return &telegram{
		httpClient: &http.Client{
			Timeout: (defaultTimeout + 10) * time.Second,
		},
		token:      token,
		timeout:    defaultTimeout,
		chats:      make(map[int64]chan string),
		userClient: userClient,
		repository: repository,
	}
}

func (t *telegram) getUpdates(ctx context.Context, offset int) chan *getUpdatesResponse {
	ch := make(chan *getUpdatesResponse, 1)

	go func() {
		defer close(ch)

		url := fmt.Sprintf(
			"https://api.telegram.org/bot%s/getUpdates?=%d&offset=%d",
			t.token,
			t.timeout,
			offset,
		)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Printf("create request error: %v", err)
			return
		}

		res, err := t.httpClient.Do(req.WithContext(ctx))
		if err != nil {
			log.Printf("make request error: %v", err)
			return
		}

		defer res.Body.Close()
		b, err := io.ReadAll(res.Body)
		if err != nil {
			if err != nil {
				log.Printf("response body read error: %v", err)
				return
			}
		}

		var data getUpdatesResponse
		err = json.Unmarshal(b, &data)
		if err != nil {
			log.Printf("unmarshal response body error: %v", err)
			return
		}

		ch <- &data
	}()

	return ch
}

var loginCommandReg = regexp.MustCompile("^/login ([A-Za-z0-9]+)$")

func (t *telegram) handleCommandLogin(ctx context.Context, chat *Chat, username string, ch chan string) error {
	err := t.userClient.GenerateOTP(ctx, username)
	if err != nil {
		log.Printf("generate otp error: %v", err)
		return err
	}
	log.Printf("waiting for OTP code...")
	err = t.sendMessage(ctx, chat, "Enter OTP code.")
	if err != nil {
		return err
	}
	otpCode := <-ch
	log.Printf("OTP CODE: %q", otpCode)
	err = t.verifyOTP(ctx, chat, username, otpCode)
	if err != nil {
		return err
	}
	log.Printf("OTP code verified")

	return nil
}

func (t *telegram) verifyOTP(ctx context.Context, chat *Chat, username string, code string) error {
	res, err := t.userClient.VerifyOTP(ctx, username, code)
	if err != nil {
		return err
	}

	err = t.repository.SetAuthToken(ctx, chat.ID, res.Token, res.UserID)
	if err != nil {
		return err
	}

	return nil
}

var subscribeCommangReg = regexp.MustCompile("^/subscribe$")

func (t *telegram) handleCommandSubscribe(ctx context.Context, chat *Chat) error {
	acc, err := t.repository.GetAccount(ctx, chat.ID)
	if err != nil {
		return err
	}

	ch, err := t.userClient.Subscribe(ctx, acc.UserID, acc.AuthToken)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				log.Printf("MESSAGE: %q", msg)
				err := t.sendMessage(ctx, chat, msg)
				if err != nil {
					log.Printf("ERR: %v", err)
				}
			}
		}
	}()

	return nil
}

func (t *telegram) handleMessage(ctx context.Context, chat *Chat, msg string, ch chan string) {
	if loginCommandReg.MatchString(msg) {
		submatches := loginCommandReg.FindStringSubmatch(msg)
		username := submatches[1]
		log.Printf("username: %q", username)
		err := t.handleCommandLogin(ctx, chat, username, ch)
		if err != nil {
			log.Printf("handler command error: %v", err)
		}
		return
	}
	if subscribeCommangReg.MatchString(msg) {
		err := t.handleCommandSubscribe(ctx, chat)
		if err != nil {
			log.Printf("handler command error: %v", err)
			return
		}
		return
	}
	log.Printf("unexpected message...%q", msg)
}

func (t *telegram) runChat(ctx context.Context, chat *Chat, ch chan string) error {
	err := t.repository.AddAccount(ctx, chat.ID)
	if err != nil {
		return err
	}

	t.sendMessage(ctx, chat, "connected")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				log.Printf("[%s] %q", chat.Username, msg)
				t.handleMessage(ctx, chat, msg, ch)
			}
		}
	}()

	return nil
}

func (t *telegram) getChatChan(ctx context.Context, chat *Chat) (chan string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	ch, ok := t.chats[chat.ID]
	if !ok {
		ch = make(chan string, 1)
		t.chats[chat.ID] = ch
		if err := t.runChat(ctx, chat, ch); err != nil {
			log.Printf("err: %v", err)
			return nil, err
		}
	}

	return ch, nil
}

func (t *telegram) sendMessage(ctx context.Context, chat *Chat, msg string) error {
	client := http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	values := map[string]interface{}{
		"chat_id": chat.ID,
		"text":    msg,
	}
	b, err := json.Marshal(values)
	if err != nil {
		return ErrInternalError
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	if res.StatusCode > 299 {
		return ErrInternalError
	}

	defer res.Body.Close()

	return nil
}

func (t *telegram) Serve(ctx context.Context) error {
	offset := 0

	for {
		ch := t.getUpdates(ctx, offset)
		timer := time.NewTimer(1 * time.Second) // To prevent spam

		select {
		case <-ctx.Done():
			return nil
		case data := <-ch:
			if data == nil {
				continue
			}
			for _, r := range data.Result {
				offset = r.UpdateID + 1
				ch, err := t.getChatChan(ctx, r.Message.Chat)
				if err != nil {
					continue
				}
				log.Printf("sending message to %s...", r.Message.Chat.Username)
				ch <- r.Message.Text
			}
		}

		<-timer.C
	}

	return nil
}

//func (t *telegram) Notify(ctx context.Context, user *user.User, msg interface{}) error {
//	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
//	//TODO implement me
//	panic("implement me")
//}
