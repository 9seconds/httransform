package events

type EventProcessor func(*Event)

type EventChannel chan<- *Event
