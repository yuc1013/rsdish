package logi

import (
	"fmt"
	"log"
	"rsdish/persist"
)

// LinkLibrary orchestrates the linking process for a single library.
// 它只处理存储卷之间的链接，根据目标卷的配置来决定链接类型。
// 如果 dryRun 为 true，它只会打印操作而不执行。
func LinkLibrary(uuid string, dryRun bool) error {
	library, ok := LogiTree[uuid]
	if !ok {
		return fmt.Errorf("library with UUID '%s' not found", uuid)
	}

	// 链接所有存储卷，互相作为源和目标。
	storages := library.Storages
	if len(storages) < 2 {
		log.Printf("Library '%s' needs at least 2 storage volumes for linking. Skipping.", uuid)
		return nil
	}

	for i, srcVol := range storages {
		for j, dstVol := range storages {
			if i == j {
				continue // 跳过链接自身
			}

			linkCreateMode := dstVol.Config.Advanced.LinkCreate
			if linkCreateMode == "none" || linkCreateMode == "" {
				log.Printf("Skipping link from '%s' to '%s' as link_create is 'none'.", srcVol.BasePath, dstVol.BasePath)
				continue
			}

			if dryRun {
				log.Printf("[DRY RUN] Would create links from '%s' to '%s' (mode: '%s')", srcVol.BasePath, dstVol.BasePath, linkCreateMode)
			} else {
				log.Printf("Creating links from '%s' to '%s' (mode: '%s')...", srcVol.BasePath, dstVol.BasePath, linkCreateMode)
				err := persist.LinkAll(srcVol.BasePath, dstVol.BasePath, linkCreateMode)
				if err != nil {
					log.Printf("Error creating links: %v", err)
				}
			}
		}
	}

	return nil
}

// LinkAllLibrary iterates through all libraries in LogiTree and calls LinkLibrary for each.
// 如果 dryRun 为 true，它只会打印操作而不执行。
func LinkAllLibrary(dryRun bool) error {
	if len(LogiTree) == 0 {
		return fmt.Errorf("no libraries found to link")
	}

	for uuid := range LogiTree {
		err := LinkLibrary(uuid, dryRun)
		if err != nil {
			log.Printf("Error processing library '%s': %v", uuid, err)
		}
	}
	return nil
}
