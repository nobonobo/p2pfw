package peerconn

import (
	"github.com/nobonobo/p2pfw/signaling"
	"github.com/nobonobo/webrtc"
)

// Connect ...
type Connect struct{}

// Kind ...
func (c *Connect) Kind() string { return "connect" }

// Offer ...
type Offer webrtc.SessionDescription

// Kind ...
func (c *Offer) Kind() string { return "offer" }

// Answer ...
type Answer webrtc.SessionDescription

// Kind ...
func (c *Answer) Kind() string { return "answer" }

// OfferCandidate ...
type OfferCandidate webrtc.IceCandidate

// Kind ...
func (c *OfferCandidate) Kind() string { return "offer-candidate" }

// OfferCompleted ...
type OfferCompleted struct{}

// Kind ...
func (c *OfferCompleted) Kind() string { return "offer-completed" }

// OfferFailed ...
type OfferFailed struct{}

// Kind ...
func (c *OfferFailed) Kind() string { return "offer-failed" }

// AnswerCandidate ...
type AnswerCandidate webrtc.IceCandidate

// Kind ...
func (c *AnswerCandidate) Kind() string { return "answer-candidate" }

// AnswerCompleted ...
type AnswerCompleted struct{}

// Kind ...
func (c *AnswerCompleted) Kind() string { return "answer-completed" }

// AnswerFailed ...
type AnswerFailed struct{}

// Kind ...
func (c *AnswerFailed) Kind() string { return "answer-failed" }

func init() {
	signaling.Register(func() signaling.Kinder { return new(Connect) })
	signaling.Register(func() signaling.Kinder { return new(Offer) })
	signaling.Register(func() signaling.Kinder { return new(OfferCandidate) })
	signaling.Register(func() signaling.Kinder { return new(OfferCompleted) })
	signaling.Register(func() signaling.Kinder { return new(OfferFailed) })
	signaling.Register(func() signaling.Kinder { return new(Answer) })
	signaling.Register(func() signaling.Kinder { return new(AnswerCandidate) })
	signaling.Register(func() signaling.Kinder { return new(AnswerCompleted) })
	signaling.Register(func() signaling.Kinder { return new(AnswerFailed) })
}
