package tron

type AccountList []Account

// Account représente un compte avec toutes ses propriétés
type Account struct {
	AccountName           string          `json:"account_name"`
	Address               string          `json:"address"`
	Balance               int64           `json:"balance"`
	CreateTime            int64           `json:"create_time"`
	LatestOperationTime   int64           `json:"latest_opration_time"`
	Allowance             int64           `json:"allowance"`
	LatestWithdrawTime    int64           `json:"latest_withdraw_time"`
	IsWitness             bool            `json:"is_witness"`
	LatestConsumeFreeTime int64           `json:"latest_consume_free_time"`
	NetWindowSize         int64           `json:"net_window_size"`
	NetWindowOptimized    bool            `json:"net_window_optimized"`
	AccountResource       AccountResource `json:"account_resource"`
	OwnerPermission       Permission      `json:"owner_permission"`
	WitnessPermission     Permission      `json:"witness_permission"`
	ActivePermission      []Permission    `json:"active_permission"`
	FrozenV2              []FrozenV2      `json:"frozenV2"`
	AssetV2               []AssetV2       `json:"assetV2"`
	FreeAssetNetUsageV2   []AssetV2       `json:"free_asset_net_usageV2"`
	AssetOptimized        bool            `json:"asset_optimized"`

	WitnessInfo *Witness `json:"-"`
}

// AccountResource représente les ressources associées au compte
type AccountResource struct {
	EnergyWindowSize      int64 `json:"energy_window_size"`
	EnergyWindowOptimized bool  `json:"energy_window_optimized"`
}

type Key struct {
	Address string `json:"address"`
	Weight  int    `json:"weight"`
}

// Permission représente une permission d'un compte
type Permission struct {
	Type           string `json:"type,omitempty"`
	ID             int    `json:"id,omitempty"`
	PermissionName string `json:"permission_name"`
	Threshold      int    `json:"threshold"`
	Operations     string `json:"operations,omitempty"`
	Keys           []Key  `json:"keys"`
}

// FrozenV2 représente un élément de la liste FrozenV2
type FrozenV2 struct {
	Amount int64  `json:"amount,omitempty"`
	Type   string `json:"type,omitempty"`
}

// AssetV2 représente une clé-valeur des actifs du compte
type AssetV2 struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

type Witness struct {
	Address        string `json:"address"`
	VoteCount      int64  `json:"voteCount"`
	URL            string `json:"url"`
	TotalProduced  int64  `json:"totalProduced,omitempty"`
	TotalMissed    int64  `json:"totalMissed,omitempty"`
	LatestBlockNum int64  `json:"latestBlockNum,omitempty"`
	LatestSlotNum  int64  `json:"latestSlotNum,omitempty"`
	IsJobs         bool   `json:"isJobs,omitempty"`

	Rank int `json:"-"`
}

// WitnessesData représente l'ensemble des témoins
type Witnesses struct {
	Witnesses []Witness `json:"witnesses"`
}

func (s *Account) IsBlockProducer() bool {
	return s.WitnessInfo.IsJobs
}

func (s AccountList) GetBlockProducerValidators() AccountList {
	var accounts AccountList
	for _, v := range s {
		if v.IsBlockProducer() {
			accounts = append(accounts, v)
		}
	}
	return accounts
}
