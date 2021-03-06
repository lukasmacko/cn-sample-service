// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package idxvpp

import (
	"errors"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/idxmap"
)

// NameToIdxDto defines the Data Transfer Object (DTO) that carries a
// mapping between a logical object name defined at Agent's Northbound
// API and an sw_if_index defined at the VPP binary API.
type NameToIdxDto struct {
	NameToIdxDtoWithoutMeta
	// Auxiliary data related to mapping
	Metadata interface{}
}

// NameToIdxDtoWithoutMeta is the part of NameToIdxDto that can be reused
// by indices with typed metadata
type NameToIdxDtoWithoutMeta struct {
	idxmap.NamedMappingDtoWithoutMeta

	Idx uint32
}

// Done is used to signal to the event producer that the event consumer
// has processed the event.
func (dto *NameToIdxDtoWithoutMeta) Done() error {
	// TODO Consumer of the channel must signal that he processed the event
	return errors.New("Unimplemented")
}

// IsDelete returns true if the mapping was deleted.
func (dto *NameToIdxDtoWithoutMeta) IsDelete() bool { // similarity to other APIs
	return dto.Del
}

// NameToIdxRW is the "owner API" to the NameToIdx registry. Using this
// API the owner adds (registers) new mappings to the registry or deletes
// (unregisters) existing mappings from the registry.
type NameToIdxRW interface {
	NameToIdx

	// RegisterName registers a new name-to-index mapping. After
	// registration, other plugins can use the "user's API" to lookup the
	// index (by providing the name) or the name by providing the index),
	// and/or be notified  when the mapping is changed. A plugins will
	// typically use the change notifications to modify the part of VPP
	// configuration that it owns and use the VPP binary API to push it
	// to VPP.
	RegisterName(name string, idx uint32, metadata interface{})

	// UnregisterName removes a mapping from the registry. Other plugins
	// can be notified and remove the relevant parts of their own respective
	// VPP configurations and use the VPP binary API to clean it up from
	// VPP.
	UnregisterName(name string) (idx uint32, metadata interface{}, exists bool)
}

// NameToIdx is the "user API" to the NameToIdx registry. It provides
// read-only access to name-to-index mappings stored in the registry. It
// is intended for plugins that need to lookup the mappings.
//
// For example, a static L2 FIB table entry refers to an underlying network
// interface, which is specified as a logical interface name at the L2
// plugin's NB API. During configuration, the LookupIdx() function must
// be called to determine the VPP if index that corresponds to the
// specified logical name.
type NameToIdx interface {
	// GetRegistryTitle returns the title assigned to the registry.
	GetRegistryTitle() string

	// LookupIdx retrieves a previously stored index for a particular
	// name. Metadata can be nil. If the 'exists' flag is set to false
	// upon return, the init value is undefined and it should be ignored.
	LookupIdx(name string) (idx uint32, metadata interface{}, exists bool)

	// LookupName retrieves a previously stored name by particular index.
	// Metadata can be nil. Name is supposed to be only when exists=false.
	//
	// Principle: A. Registry stores mappings between names and indexes. API can optionally attach metadata to a particular name. TBD index can be 0... B. Metadata is needed for example in the ifplugin. This metadata is used in the following scenarios:
	// - for caching of data (even data that belongs to a different agent)
	// - for remembering the last processed object
	// - for indexing the BD to which a particular interface belongs to (see bd_configurator or fib_configurator)
	LookupName(idx uint32) (name string, metadata interface{}, exists bool)

	// LookupNameByMetadata returns all indexes that contains particular meta field with provided value.
	LookupNameByMetadata(key string, value string) []string

	// ListNames returns all names in the mapping
	ListNames() (names []string)

	// Watch subscribes to watching changes to name-to-index
	// mappings.
	// Watching NameToIndex mapping is critical for performance. You could delay something if you are very slowly handling these events. The idea of this text was to prepare disclaimer for that.
	//
	// Example:
	//
	//    func (plugin *Plugin) watchEvents(ctx context.Context) {
	//       Watch(PluginID, plugin.isubscriberChan)
	//        ...
	//		 select {
	// 		    case ifIdxEv := <-plugin.isubscriberChan:
	//
	//			if ifIdxEv.IsDelete() {
	//				plugin.ResolveDeletedInterface(ifIdxEv.Name)
	//			} else {
	//				plugin.ResolveDeletedInterface(ifIdxEv.Name)
	//			}
	//			ifIdxEv.Done()
	//       ...
	//    }
	Watch(subscriber core.PluginName, callback func(NameToIdxDto))
}
