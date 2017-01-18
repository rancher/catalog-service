package service

/*func filterByMinimumRancherVersion(rancherVersion string, template *model.Template) (map[string]string, error) {
        copyOfversionLinks := make(map[string]string)

        vB, err := getSemVersion(rancherVersion)
        if err != nil {
                log.Errorf("Error loading the passed filter minimumRancherVersion_lte with semver %s", err.Error())
                return copyOfversionLinks, err
        }

        for templateVersion, minRancherVersion := range template.TemplateVersionRancherVersion {
                if minRancherVersion != "" {
                        vA, err := getSemVersion(minRancherVersion)
                        if err != nil {
                                log.Errorf("Error loading version with semver %s", err.Error())
                                continue
                        }

                        if minRancherVersion == rancherVersion || vA.LT(*vB) {
                                //this template version passes the filter
                                if template.VersionLinks[templateVersion] != "" {
                                        copyOfversionLinks[templateVersion] = template.VersionLinks[templateVersion]
                                }
                        }
                } else {
                        //no min rancher version specified, so this template works with any rancher version
                        if template.VersionLinks[templateVersion] != "" {
                                copyOfversionLinks[templateVersion] = template.VersionLinks[templateVersion]
                        }
                }
        }

        return copyOfversionLinks, nil
}

func filterByMaximumRancherVersion(rancherVersion string, template *model.Template) (map[string]string, error) {
        copyOfversionLinks := make(map[string]string)

        vB, err := getSemVersion(rancherVersion)
        if err != nil {
                log.Errorf("Error loading the passed filter maximumRancherVersion_gte with semver %s", err.Error())
                return copyOfversionLinks, err
        }

        for templateVersion, maxRancherVersion := range template.TemplateVersionRancherVersionGte {
                if maxRancherVersion != "" {
                        vA, err := getSemVersion(maxRancherVersion)
                        if err != nil {
                                log.Errorf("Error loading version with semver %s", err.Error())
                                continue
                        }

                        if maxRancherVersion == rancherVersion || vA.GT(*vB) {
                                //this template version passes the filter
                                if template.VersionLinks[templateVersion] != "" {
                                        copyOfversionLinks[templateVersion] = template.VersionLinks[templateVersion]
                                }
                        }
                } else {
                        //no max rancher version specified, so this template works with any rancher version
                        if template.VersionLinks[templateVersion] != "" {
                                copyOfversionLinks[templateVersion] = template.VersionLinks[templateVersion]
                        }
                }
        }

        return copyOfversionLinks, nil
}

func getSemVersion(versionStr string) (*semver.Version, error) {
        versionStr = re.ReplaceAllString(versionStr, "$1")

        semVersion, err := semver.Make(versionStr)
        if err != nil {
                log.Errorf("Error %v loading semver for version string %s", err.Error(), versionStr)
                return nil, err
        }
        return &semVersion, nil
}

func isMinRancherVersionLTE(templateMinRancherVersion string, rancherVersion string) (bool, error) {
        vA, err := getSemVersion(templateMinRancherVersion)
        if err != nil {
                log.Errorf("Error loading template minRancherVersion %s with semver %s", templateMinRancherVersion, err.Error())
                return false, err
        }

        vB, err := getSemVersion(rancherVersion)
        if err != nil {
                log.Errorf("Error loading the passed filter minimumRancherVersion_lte %s with semver %s", rancherVersion, err.Error())
                return false, err
        }

        if templateMinRancherVersion == rancherVersion || vA.LT(*vB) {
                return true, nil
        }
        return false, nil
}

func isMaxRancherVersionGTE(templateMaxRancherVersion string, rancherVersion string) (bool, error) {
        vA, err := getSemVersion(templateMaxRancherVersion)
        if err != nil {
                log.Errorf("Error loading template maxRancherVersion %s with semver %s", templateMaxRancherVersion, err.Error())
                return false, err
        }

        vB, err := getSemVersion(rancherVersion)
        if err != nil {
                log.Errorf("Error loading the passed filter maximumRancherVersion_gte %s with semver %s", rancherVersion, err.Error())
                return false, err
        }

        if templateMaxRancherVersion == rancherVersion || vA.GT(*vB) {
                return true, nil
        }
        return false, nil
}*/
