package apiClient

// Resource
type CreateResourceResponse struct {
	Resource struct {
		Turbot TurbotResourceMetadata
	}
}

type UpdateResourceResponse struct {
	Resource struct {
		Turbot TurbotResourceMetadata
	}
}

// note: the Resource property is just an interface{} - this is mapped manually into a Resource object,
// rather than unmarshalled. This is to allow for dynamic data types, while always having the Turbot property
type ReadResourceResponse struct {
	Resource interface{}
}

type ReadResourceListResponse struct {
	ResourceList struct {
		Items []Resource
	}
}

type Resource struct {
	Turbot TurbotResourceMetadata
	Data   map[string]interface{}
}

type ReadSerializableResourceResponse struct {
	Resource struct {
		Data   map[string]interface{}
		Turbot map[string]interface{}
		Tags   map[string]string
		Akas   []string
	}
}

type SerializableResource struct {
	Data     string
	Metadata string
	Tags     map[string]string
	Akas     []string
	Turbot   map[string]string
}

// Validation response
type ValidationResponse struct {
	Schema struct {
		QueryType struct {
			Name string
		}
	}
}

// is the validation response successful?
func (response *ValidationResponse) isValid() bool {
	return response.Schema.QueryType.Name == "Query"
}

// ApiResponse
// used to unmarshall API error responses
type ApiResponse struct {
	Errors []Error
}

type Error struct {
	Message string
}

// PolicySetting
type PolicySettingResponse struct {
	PolicySetting PolicySetting
}

type FindPolicySettingResponse struct {
	PolicySettings struct {
		Items []PolicySetting
	}
}

type PolicySetting struct {
	Value              interface{}
	ValueSource        string
	Default            bool
	Precedence         string
	Template           string
	TemplateInput      string
	Input              string
	Note               string
	ValidFromTimestamp string
	ValidToTimestamp   string
	Turbot             TurbotPolicyMetadata
}

// PolicyValue
type PolicyValueResponse struct {
	PolicyValue PolicyValue
}

type PolicyValue struct {
	Value      interface{}
	Precedence string
	State      string
	Reason     string
	Details    string
	Setting    PolicySetting
	Turbot     TurbotPolicyMetadata
}

// Mod
type InstallModResponse struct {
	Mod InstallModData
}

type InstallModData struct {
	Build  string
	Turbot TurbotResourceMetadata
}

type ReadModResponse struct {
	Mod Mod
}

type ModRegistryVersion struct {
	Status  string
	Version string
}

type ModVersionResponse struct {
	Versions struct {
		Items []ModRegistryVersion
	}
}

type UninstallModResponse struct {
	UninstallMod struct {
		Success bool
	}
}

type Mod struct {
	Org     string
	Mod     string
	Version string
	Parent  string
	Uri     string
}

// Grant
type CreateGrantResponse struct {
	Grants struct {
		Turbot TurbotGrantMetadata
	}
}

type ReadGrantResponse struct {
	Grant Grant
}

type Grant struct {
	Turbot            TurbotGrantMetadata
	PermissionTypeId  string
	PermissionLevelId string
}

// Active Grant
type ActivateGrantResponse struct {
	GrantActivate struct {
		Turbot TurbotActiveGrantMetadata
	}
}

type ReadActiveGrantResponse struct {
	ActiveGrant ActiveGrant
}

type ActiveGrant struct {
	Turbot TurbotActiveGrantMetadata
}

// Folder
type FolderResponse struct {
	Resource Folder
}

type Folder struct {
	Turbot      TurbotResourceMetadata
	Title       string
	Description string
	Parent      string
}

// Profile
type ProfileResponse struct {
	Resource Profile
}

type Profile struct {
	Turbot          TurbotResourceMetadata
	Title           string
	Parent          string
	Status          string
	Email           string
	GivenName       string
	DisplayName     string
	FamilyName      string
	DirectoryPoolId string
	ProfileId       string
}

// Smart folder

type SmartFolderResponse struct {
	SmartFolder SmartFolder
}

type SmartFolder struct {
	Turbot            TurbotResourceMetadata
	Title             string
	Description       string
	Filters           []string
	Parent            string
	AttachedResources struct {
		Items []struct {
			Turbot TurbotResourceMetadata
		}
	}
}

// Smart folder attachment
type SmartFolderAttachment struct {
	Turbot      TurbotResourceMetadata
	Title       string
	Description string
	Filters     map[string]interface{}
	Parent      string
}

type CreateSmartFolderAttachResponse struct {
	SmartFolderAttach struct {
		Turbot TurbotResourceMetadata
	}
}

// Local directory
type LocalDirectoryResponse struct {
	Resource LocalDirectory
}

type LocalDirectory struct {
	Turbot            TurbotResourceMetadata
	Title             string
	Description       string
	Parent            string
	Status            string
	DirectoryType     string
	ProfileIdTemplate string
}

// Local directory user
type LocalDirectoryUserResponse struct {
	Resource LocalDirectoryUser
}

type LocalDirectoryUser struct {
	Turbot      TurbotResourceMetadata
	Parent      string
	Title       string
	Email       string
	Status      string
	DisplayName string
	GivenName   string
	MiddleName  string
	FamilyName  string
	Picture     string
}

// Saml directory
type SamlDirectoryResponse struct {
	Resource SamlDirectory
}

type SamlDirectory struct {
	Turbot            TurbotResourceMetadata
	Title             string
	Description       string
	Parent            string
	Status            string
	DirectoryType     string
	ProfileIdTemplate string
	EntryPoint        string
	Certificate       string
}

// Google directory
type ReadGoogleDirectoryResponse struct {
	Directory GoogleDirectory
}

type GoogleDirectory struct {
	Turbot            TurbotResourceMetadata
	Parent            string
	Title             string
	ProfileIdTemplate string
	Description       string
	Status            string
	DirectoryType     string
	ClientID          string
	ClientSecret      string
	PoolId            string
	GroupIdTemplate   string
	LoginNameTemplate string
	HostedName        string
}

// Metadata
type TurbotResourceMetadata struct {
	Id                string
	ParentId          string
	Akas              []string
	Custom            map[string]interface{}
	Metadata          map[string]interface{}
	Tags              map[string]string
	Title             string
	VersionId         string
	ActorIdentityId   string
	ActorPersonaId    string
	ActorRoleId       string
	ResourceParentAka string
	CreateTimestamp   string
	DeleteTimestamp   string
	UpdateTimestamp   string
	Path              string
	ResourceGroupIds  []string
	ResourceTypeId    string
	State             string
	Terraform         map[string]interface{}
}

type TurbotPolicyMetadata struct {
	Id         string
	ParentId   string
	ResourceId string
	Akas       []string
}

type TurbotGrantMetadata struct {
	Id         string
	ProfileId  string
	ResourceId string
}

type TurbotActiveGrantMetadata struct {
	Id         string
	GrantId    string
	ResourceId string
}
