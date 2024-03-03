package crypto

import (
	"encoding/json"
	"fmt"
	"strconv"

	"server/abe"
)

func Encrypt(dep, level int, file []byte) ([]byte, []byte, []byte, error) {
	attribDep, attribLevel, err := createAttribs(dep, level)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to createAttribs: %w", err)
	}
	maabe := abe.NewMAABE()
	depAuth, levelAuth, err := createAuthorities(attribDep, attribLevel, maabe)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to createAuthorities: %w", err)
	}

	policy, err := createPolicy(dep, level)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to createPolicy: %w", err)
	}

	msp, err := abe.BooleanToMSP(policy, false)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to BooleanToMSP: %w", err)
	}

	// define the set of all public keys we use
	pks := []*abe.MAABEPubKey{depAuth.PubKeys(), levelAuth.PubKeys()}

	ct, err := maabe.Encrypt(string(file), msp, pks)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to Encrypt: %w", err)
	}

	cipherRaw, err := json.Marshal(ct)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to Marshal: %w", err)
	}

	levelAuthRaw, err := json.Marshal(levelAuth)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to Marshal level auth: %w", err)
	}

	depAuthRaw, err := json.Marshal(depAuth)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to Marshal dep auth: %w", err)
	}

	return cipherRaw, depAuthRaw, levelAuthRaw, nil
}

func Decrypt(dep, level int, file []byte, depAuthRaw, levelAuthRaw []byte) ([]byte, error) {
	var (
		levelAuth = new(abe.MAABEAuth)
		depAuth   = new(abe.MAABEAuth)
	)

	if err := json.Unmarshal(levelAuthRaw, levelAuth); err != nil {
		return nil, fmt.Errorf("failed to Unmarshal level auth: %w", err)
	}

	if err := json.Unmarshal(depAuthRaw, depAuth); err != nil {
		return nil, fmt.Errorf("failed to Unmarshal dep auth: %w", err)
	}

	attribDep, attribLevel, err := createAttribs(dep, level)
	if err != nil {
		return nil, fmt.Errorf("failed to createAttribs: %w", err)
	}

	depKeys, err := depAuth.GenerateAttribKeys("gid", attribDep)
	if err != nil {
		return nil, fmt.Errorf("failed to GenerateAttribKeys: %w", err)
	}

	levelKeys, err := levelAuth.GenerateAttribKeys("gid", attribLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to GenerateAttribKeys: %w", err)
	}

	if len(depKeys) != 1 {
		return nil, fmt.Errorf("invalid number of department keys: %w", err)
	}

	ks := []*abe.MAABEKey{depKeys[0]}
	for _, key := range levelKeys {
		ks = append(ks, key)
	}

	var cipher = new(abe.MAABECipher)
	if err = json.Unmarshal(file, cipher); err != nil {
		return nil, fmt.Errorf("failed to Unmarshal: %w", err)
	}

	decrypted, err := abe.NewMAABE().Decrypt(cipher, ks)
	if err != nil {
		return nil, fmt.Errorf("failed to Decrypt: %w", err)
	}

	return []byte(decrypted), nil
}

const (
	securityLevelNotSecret = iota
	securityLevelForOfficialUse
	securityLevelSecretly
	securityLevelAbsolutelySecretly
	securityLevelSpacialImportance
)

func createAttribs(department, securityLevel int) ([]string, []string, error) {
	depAuthorityRaw := []string{"department:" + strconv.Itoa(department)}

	var levelAuthorityRaw []string
	switch securityLevel {
	case 0:
		levelAuthorityRaw = []string{"level:" + strconv.Itoa(securityLevelNotSecret), "level:" + strconv.Itoa(securityLevelForOfficialUse), "level:" + strconv.Itoa(securityLevelSecretly), "level:" + strconv.Itoa(securityLevelAbsolutelySecretly), "level:" + strconv.Itoa(securityLevelSpacialImportance)}

	case 1:
		levelAuthorityRaw = []string{"level:" + strconv.Itoa(securityLevelForOfficialUse), "level:" + strconv.Itoa(securityLevelSecretly), "level:" + strconv.Itoa(securityLevelAbsolutelySecretly), "level:" + strconv.Itoa(securityLevelSpacialImportance)}

	case 2:
		levelAuthorityRaw = []string{"level:" + strconv.Itoa(securityLevelSecretly), "level:" + strconv.Itoa(securityLevelAbsolutelySecretly), "level:" + strconv.Itoa(securityLevelSpacialImportance)}

	case 3:
		levelAuthorityRaw = []string{"level:" + strconv.Itoa(securityLevelAbsolutelySecretly), "level:" + strconv.Itoa(securityLevelSpacialImportance)}

	case 4:
		levelAuthorityRaw = []string{"level:" + strconv.Itoa(securityLevelSpacialImportance)}
	default:
		return nil, nil, fmt.Errorf("unknown policy")
	}

	return depAuthorityRaw, levelAuthorityRaw, nil
}

func createAuthorities(depAttrib, levelAttrib []string, maabe *abe.MAABE) (*abe.MAABEAuth, *abe.MAABEAuth, error) {
	depAuth, err := maabe.NewMAABEAuth("department", depAttrib)
	if err != nil {
		return nil, nil, fmt.Errorf("failed generation authority %s: %w", "department", err)
	}

	levelAuth, err := maabe.NewMAABEAuth("level", levelAttrib)
	if err != nil {
		return nil, nil, fmt.Errorf("failed generation authority %s: %w", "level", err)
	}

	return depAuth, levelAuth, nil
}

func createPolicy(department, securityLevel int) (string, error) {
	policy := fmt.Sprintf("department:%d AND", department)
	switch securityLevel {
	case 0:
		policy = fmt.Sprintf("%s (%s OR %s OR %s OR %s OR %s)", policy,
			fmt.Sprintf("level:%d", securityLevelNotSecret),
			fmt.Sprintf("level:%d", securityLevelForOfficialUse),
			fmt.Sprintf("level:%d", securityLevelSecretly),
			fmt.Sprintf("level:%d", securityLevelAbsolutelySecretly),
			fmt.Sprintf("level:%d", securityLevelSpacialImportance))
	case 1:
		policy = fmt.Sprintf("%s (%s OR %s OR %s OR %s)", policy,
			fmt.Sprintf("level:%d", securityLevelForOfficialUse),
			fmt.Sprintf("level:%d", securityLevelSecretly),
			fmt.Sprintf("level:%d", securityLevelAbsolutelySecretly),
			fmt.Sprintf("level:%d", securityLevelSpacialImportance))
	case 2:
		policy = fmt.Sprintf("%s (%s OR %s OR %s)", policy,
			fmt.Sprintf("level:%d", securityLevelSecretly),
			fmt.Sprintf("level:%d", securityLevelAbsolutelySecretly),
			fmt.Sprintf("level:%d", securityLevelSpacialImportance))
	case 3:
		policy = fmt.Sprintf("%s (%s OR %s)", policy,
			fmt.Sprintf("level:%d", securityLevelAbsolutelySecretly),
			fmt.Sprintf("level:%d", securityLevelSpacialImportance))
	case 4:
		policy = fmt.Sprintf("%s %s", policy,
			fmt.Sprintf("level:%d", securityLevelSpacialImportance))
	default:
		return "", fmt.Errorf("unknown policy")
	}

	return policy, nil
}
