package database

import (
	"fmt"
	"log"
	"os"

	"../crawler"

	bolt "go.etcd.io/bbolt"
)

var pageInfo *bolt.DB
var pageInfoBuck string = "pageInfoBuck"
var parentChildBuck string = "parentChildBuck"
var childParentBuck string = "childParentBuck"

func openPageInfoDb() {
  var err error
  pageInfo, err = bolt.Open("db" + string(os.PathSeparator) + "pageInfo.db", 0700, nil)

  if err != nil {
    log.Fatal(err)
  }

  err = pageInfo.Update(func(tx *bolt.Tx) error {
    _, err := tx.CreateBucketIfNotExists([]byte(pageInfoBuck))
    if err != nil {
      return fmt.Errorf("pageInfo create first bucket: %s", err)
    }

    _, err = tx.CreateBucketIfNotExists([]byte(parentChildBuck))
    if err != nil {
      return fmt.Errorf("pageInfo create second bucket: %s", err)
    }

    _, err = tx.CreateBucketIfNotExists([]byte(childParentBuck))
    if err != nil {
      return fmt.Errorf("pageInfo create third bucket: %s", err)
    }

    return nil
  })

  if err != nil {
    log.Fatal(err)
  }
}

func closePageInfoDb() {
  pageInfo.Close()
}

// parse all the info on the given page, including all it's child links
func parseAllInfo(page *crawler.Page) {
  var info []string
  info = append(info, page.GetTitle())
  info = append(info, page.GetURL())
  info = append(info, page.GetLastModified())
  info = append(info, page.GetSize())

  value:= StringToByte(info)
  err := pageInfo.Update(func(tx *bolt.Tx) error {
    // store the info in pageInfoBucket
    pageInfoBucket := tx.Bucket([]byte(pageInfoBuck))
    pageId := IntToByte(GetPageId(page.GetURL()))
    err := pageInfoBucket.Put(pageId, value)
    if err != nil {
      return fmt.Errorf("Error in pageInfo: parseAllInfo insert pageInfo error: %s", err)
    }

    // store the parent child relationship in parentChildBucket
    // TODO convert the child values to pageId
    parentChildBucket := tx.Bucket([]byte(parentChildBuck))
    info = page.GetChildrenURL()
    value = StringToByte(info)
    err = parentChildBucket.Put(pageId, value)
    if err != nil {
      return fmt.Errorf("Error in pageInfo: parseAllInfo insert parentChild error: %s", err)
    }

    // TODO store the child parent relationship in childParentBucket
    // check first if GetParentURL is not empty, if not empty, handle
    childParentBucket := tx.Bucket([]byte(childParentBuck))
    info = page.GetParentURL()
    
    // handle info nil
    if info != nil {
      // check first if contents of db are empty for that particular pageid
      parentUrlInDbByte := childParentBucket.Get(pageId)
      if parentUrlInDbByte != nil {
        parentUrlInDbString := ByteToString(parentUrlInDbByte)
        toBeInserted := parentUrlInDbString
        for _, url := range info {
          existInDb := false
          for _, parent := range parentUrlInDbString {
            if url == parent {
              existInDb = true
              break
            }
          }
          if existInDb == false {
            toBeInserted = append(toBeInserted, url)
            // fmt.Println("toBeInserted: ", toBeInserted)
          }
        }
        if toBeInserted != nil {
          err = childParentBucket.Put(pageId, StringToByte(toBeInserted))
          if err != nil {
            return fmt.Errorf("Error in pageInfo: childToParent error: %s", err)
          }
        }
      } else {
        err = childParentBucket.Put(pageId, StringToByte(info))
        if err != nil {
          return fmt.Errorf("Error in pageInfo: childToParent error: %s", err)
        }
      }
    }

    return nil
  })
  if err != nil {
    log.Fatal(err)
  }
}

// given a page url, return the child []string
func FindChild(url string) (ret []string) {
  pageId := IntToByte(GetPageId(url))
  err := pageInfo.View(func(tx *bolt.Tx) error {
    parentChildBucket := tx.Bucket([]byte(parentChildBuck))
    value := parentChildBucket.Get(pageId)
    if value != nil {
      ret = ByteToString(parentChildBucket.Get(value))
    } else {
      ret = nil
    }

    return nil
  })
  if err != nil {
    log.Fatal(err)
  }
  
  return ret
}

// given a page url, return the parent []string
func FindParent(url string) (ret []string) {
  pageId := IntToByte(GetPageId(url))
  err := pageInfo.View(func(tx *bolt.Tx) error {
    childParentBucket := tx.Bucket([]byte(childParentBuck))
    value := childParentBucket.Get(pageId)
    if value != nil {
      ret = ByteToString(value)
    } else {
      ret = nil
    }
    return nil
  })
  if err != nil {
    log.Fatal(err)
  }

  return ret
}

// print pageInfoDb in human readable format
func PrintPageInfoDb() {
  pageInfo.View(func(tx *bolt.Tx) error {
    pageInfoBucket := tx.Bucket([]byte(pageInfoBuck))
    c := pageInfoBucket.Cursor()

    fmt.Println("pageInfoBucket")
    for k, v := c.First(); k != nil; k, v = c.Next() {
      if v != nil {
        fmt.Println("key: ", ByteToInt(k), "value: ", ByteToString(v))
      }
    }

    fmt.Println("parentChildBucket")
    parentChildBucket := tx.Bucket([]byte(parentChildBuck))
    c = parentChildBucket.Cursor()

    for k, v := c.First(); k != nil; k, v = c.Next() {
      if v != nil {
        fmt.Println("key: ", ByteToInt(k), "value: ", ByteToString(v))
      }
    }

    fmt.Println("childParentBucket")
    childParentBucket := tx.Bucket([]byte(childParentBuck))
    c = childParentBucket.Cursor()

    for k, v := c.First(); k != nil; k, v = c.Next() {
      if v != nil {
        fmt.Println("key: ", ByteToInt(k), "value: ", ByteToString(v))
      }
    }

    return nil
  })
}