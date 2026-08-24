// Package web embeds the four ConfigHub control pages.
package web

import _ "embed"

//go:embed items.html
var ItemsHTML string

//go:embed releases.html
var ReleasesHTML string

//go:embed subscriptions.html
var SubscriptionsHTML string

//go:embed audit.html
var AuditHTML string
