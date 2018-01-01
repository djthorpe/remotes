/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2016-2018
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package sony

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Sony Configuration
type Sony struct {
}

type sony struct {
	log gopi.Logger
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Sony) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<remotes.Sony.Open>{ }")

	this := new(sony)
	this.log = log

	return this, nil
}

func (this *sony) Close() error {
	this.log.Debug2("<emotes.Sony.Close>{ }")
	return nil
}
