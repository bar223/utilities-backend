package ds

// Сериализаторы — структуры, через которые модели переводятся в JSON для
// клиента и обратно. Системные поля (id, status, creator, даты) в
// запросах отсутствуют: клиент не может их изменить, они вычисляются на
// бэкенде.

// UserSerializer — создатель услуги, вложенный в ответ ленты.
type UserSerializer struct {
	ID    uint   `json:"id"`
	Login string `json:"login"`
}

// UtilityListItem — услуга в списке. isMine: 1 — создатель услуги
// совпадает с текущим пользователем, 0 — нет.
type UtilityListItem struct {
	Utility
	IsMine     int   `json:"isMine"`
	LikesCount int64 `json:"likesCount"`
}

// UtilityFeedItem — услуга в ленте с вложенным создателем (вложенная
// сериализация). liked: 1 — текущий пользователь лайкнул услугу, 0 — нет.
type UtilityFeedItem struct {
	Utility
	Creator    UserSerializer `json:"creator"`
	Liked      int            `json:"liked"`
	LikesCount int64          `json:"likesCount"`
}

// UtilityLikeResponse — ответ на лайк: состояние после запроса.
type UtilityLikeResponse struct {
	Liked      int   `json:"liked"`
	LikesCount int64 `json:"likesCount"`
}

// PublishUtilityRequest — тело PUT публикации: только поля, которые
// заполняет пользователь.
type PublishUtilityRequest struct {
	Description     string  `json:"description"`
	TariffRate      float64 `json:"tariffRate" binding:"required,gt=0"`
	ConsumptionNorm float64 `json:"consumptionNorm" binding:"required,gt=0"`
	MeasurementUnit string  `json:"measurementUnit" binding:"required"`
}

// LikeRequest — тело POST лайка: 0 отменяет лайк, 1 ставит.
type LikeRequest struct {
	Value *int `json:"value" binding:"required,oneof=0 1"`
}

// RegisterRequest — тело POST регистрации и аутентификации.
type RegisterRequest struct {
	Login    string `json:"login" example:"ivanov"`
	Password string `json:"password" example:"secret"`
}

// LoginResponse — ответ аутентификации: JWT и его срок жизни в секундах.
type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType" example:"Bearer"`
	ExpiresIn   int64  `json:"expiresIn" example:"3600"`
	Login       string `json:"login" example:"ivanov"`
}

// BoolToInt переводит признак в 0/1 для JSON-ответа.
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
