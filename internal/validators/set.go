package validators

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/google/uuid"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"github.com/ttacon/libphonenumber"
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

type ValidatorSet struct {
	services common.Services
	provider.LMT
}

var (
	zipUsaRegexp = regexp.MustCompile("^[0-9]{5}(?:-[0-9]{4})?$")
	nameRegexp   = regexp.MustCompile("^[\\p{L}\\p{M} \\-\\']+$")
	cityRegexp   = regexp.MustCompile("^[\\p{L}\\p{M} \\-\\.]+$")
	localeRegexp = regexp.MustCompile("^[a-z]{2}-[A-Z]{2,10}$")
)

// PhoneValidator
func (v *ValidatorSet) PhoneValidator(fl validator.FieldLevel) bool {
	_, err := libphonenumber.Parse(fl.Field().String(), "US")
	return err == nil
}

// UuidValidator
func (v *ValidatorSet) UuidValidator(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

// ZipUsaValidator
func (v *ValidatorSet) ZipUsaValidator(fl validator.FieldLevel) bool {
	return zipUsaRegexp.MatchString(fl.Field().String())
}

// NameValidator
func (v *ValidatorSet) NameValidator(fl validator.FieldLevel) bool {
	return nameRegexp.MatchString(fl.Field().String())
}

// CityValidator
func (v *ValidatorSet) CityValidator(fl validator.FieldLevel) bool {
	return cityRegexp.MatchString(fl.Field().String())
}

// User locale validator
func (v *ValidatorSet) UserLocaleValidator(fl validator.FieldLevel) bool {
	return localeRegexp.MatchString(fl.Field().String())
}

// New
func New(services common.Services, set provider.AwareSet) *ValidatorSet {
	set.Logger = set.Logger.WithFields(logger.Fields{"service": Prefix})
	return &ValidatorSet{services: services, LMT: &set}
}
