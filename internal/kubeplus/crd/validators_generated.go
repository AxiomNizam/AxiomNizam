package crd

func ValidateTemplate001(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 001")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate002(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 002")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate003(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 003")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate004(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 004")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate005(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 005")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate006(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 006")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate007(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 007")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate008(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 008")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate009(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 009")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate010(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 010")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate011(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 011")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate012(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 012")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate013(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 013")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate014(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 014")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate015(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 015")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate016(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 016")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate017(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 017")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate018(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 018")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate019(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 019")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate020(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 020")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate021(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 021")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate022(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 022")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate023(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 023")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate024(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 024")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate025(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 025")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate026(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 026")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate027(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 027")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate028(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 028")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate029(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 029")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate030(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 030")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate031(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 031")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate032(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 032")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate033(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 033")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate034(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 034")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate035(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 035")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate036(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 036")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate037(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 037")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate038(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 038")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate039(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 039")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate040(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 040")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate041(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 041")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate042(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 042")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate043(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 043")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate044(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 044")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate045(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 045")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate046(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 046")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate047(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 047")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate048(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 048")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate049(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 049")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate050(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 050")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate051(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 051")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate052(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 052")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate053(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 053")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate054(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 054")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate055(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 055")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate056(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 056")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate057(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 057")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate058(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 058")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate059(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 059")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate060(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 060")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate061(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 061")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate062(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 062")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate063(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 063")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate064(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 064")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate065(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 065")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate066(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 066")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate067(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 067")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate068(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 068")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate069(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 069")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate070(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 070")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate071(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 071")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate072(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 072")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate073(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 073")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate074(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 074")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate075(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 075")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate076(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 076")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate077(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 077")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate078(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 078")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate079(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 079")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate080(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 080")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate081(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 081")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate082(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 082")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate083(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 083")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate084(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 084")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate085(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 085")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate086(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 086")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate087(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 087")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate088(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 088")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate089(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 089")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate090(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 090")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate091(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 091")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate092(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 092")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate093(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 093")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate094(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 094")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate095(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 095")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate096(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 096")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate097(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 097")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate098(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 098")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate099(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 099")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate100(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 100")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate101(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 101")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate102(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 102")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate103(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 103")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate104(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 104")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate105(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 105")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate106(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 106")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate107(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 107")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate108(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 108")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate109(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 109")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate110(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 110")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate111(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 111")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate112(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 112")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate113(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 113")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate114(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 114")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate115(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 115")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate116(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 116")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate117(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 117")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate118(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 118")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate119(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 119")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate120(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 120")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate121(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 121")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate122(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 122")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate123(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 123")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate124(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 124")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate125(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 125")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate126(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 126")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate127(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 127")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate128(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 128")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate129(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 129")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate130(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 130")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate131(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 131")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate132(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 132")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate133(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 133")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate134(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 134")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate135(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 135")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate136(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 136")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate137(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 137")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate138(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 138")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate139(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 139")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate140(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 140")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate141(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 141")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate142(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 142")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate143(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 143")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate144(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 144")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate145(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 145")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate146(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 146")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate147(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 147")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate148(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 148")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate149(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 149")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate150(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 150")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate151(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 151")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate152(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 152")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate153(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 153")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate154(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 154")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate155(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 155")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate156(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 156")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate157(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 157")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate158(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 158")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate159(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 159")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate160(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 160")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate161(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 161")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate162(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 162")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate163(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 163")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate164(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 164")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate165(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 165")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate166(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 166")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate167(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 167")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate168(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 168")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate169(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 169")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate170(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 170")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate171(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 171")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate172(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 172")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate173(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 173")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate174(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 174")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate175(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 175")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate176(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 176")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate177(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 177")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate178(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 178")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate179(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 179")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate180(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 180")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate181(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 181")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate182(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 182")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate183(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 183")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate184(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 184")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate185(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 185")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate186(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 186")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate187(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 187")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate188(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 188")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate189(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 189")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate190(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 190")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate191(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 191")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate192(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 192")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate193(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 193")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate194(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 194")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate195(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 195")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate196(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 196")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate197(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 197")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate198(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 198")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate199(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 199")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate200(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 200")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate201(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 201")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate202(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 202")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate203(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 203")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate204(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 204")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate205(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 205")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate206(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 206")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate207(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 207")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate208(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 208")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate209(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 209")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate210(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 210")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate211(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 211")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate212(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 212")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate213(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 213")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate214(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 214")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate215(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 215")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate216(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 216")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate217(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 217")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate218(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 218")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate219(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 219")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}

func ValidateTemplate220(spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	if spec == nil {
		return ValidationResult{Valid: false, Errors: []string{"spec cannot be nil"}, Warnings: []string{}}
	}
	if _, ok := spec["name"]; !ok {
		res.Valid = false
		res.Errors = append(res.Errors, "missing name for template 220")
	}
	if v, ok := spec["replicas"]; ok {
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			res.Valid = false
			res.Errors = append(res.Errors, "replicas must be numeric")
		}
	} else {
		res.Warnings = append(res.Warnings, "replicas not set, defaulting to 1")
	}
	return res
}
