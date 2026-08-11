package storage

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"apd/internal/wiki"
	"gopkg.in/yaml.v3"
)

type journal struct {
	Raw         string `json:"raw"`
	Receipt     string `json:"receipt"`
	RawHash     string `json:"raw_hash"`
	ReceiptHash string `json:"receipt_hash"`
}

func publishReceipt(workspace, receiptPath, rawPath string, receipt Receipt, raw []byte) error {
	transaction := filepath.Join(workspace, ".apd", "transactions", receipt.ID+".json")
	data, err := yaml.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("marshal receipt: %w", err)
	}
	entry, _ := json.Marshal(journal{Raw: rawPath, Receipt: receiptPath, RawHash: hash(raw), ReceiptHash: hash(data)})
	if err := atomicWrite(transaction, entry, 0o600); err != nil {
		return err
	}
	if err := exclusiveWrite(rawPath, raw, 0o600); err != nil {
		return err
	}
	if err := exclusiveWrite(receiptPath, data, 0o600); err != nil {
		return err
	}
	return removeAndSync(transaction)
}

func recoverJournals(workspace string) error {
	entries, err := os.ReadDir(filepath.Join(workspace, ".apd", "transactions"))
	if err != nil {
		return fmt.Errorf("read transactions: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(workspace, ".apd", "transactions", entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var j journal
		if json.Unmarshal(data, &j) != nil || !transactionPaths(workspace, j) {
			return fmt.Errorf("invalid transaction %q", path)
		}
		raw, rawErr := os.ReadFile(j.Raw)
		receipt, receiptErr := os.ReadFile(j.Receipt)
		if rawErr == nil && receiptErr == nil && hash(raw) == j.RawHash && hash(receipt) == j.ReceiptHash {
			if err := removeAndSync(path); err != nil {
				return err
			}
			continue
		}
		if err := removeAndSync(j.Raw); err != nil {
			return err
		}
		if err := removeAndSync(j.Receipt); err != nil {
			return err
		}
		if err := removeAndSync(path); err != nil {
			return err
		}
	}
	return nil
}

func transactionPaths(workspace string, j journal) bool {
	return j.RawHash != "" && j.ReceiptHash != "" &&
		contained(filepath.Join(workspace, "raw"), j.Raw) &&
		contained(filepath.Join(workspace, ".apd", "sources"), j.Receipt)
}

func loadReceipt(path string) (Receipt, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Receipt{}, err
	}
	var receipt Receipt
	if err := yaml.Unmarshal(data, &receipt); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func loadWorkspace(path string) (wiki.Workspace, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return wiki.Workspace{}, err
	}
	var workspace wiki.Workspace
	if err := yaml.Unmarshal(data, &workspace); err != nil {
		return wiki.Workspace{}, fmt.Errorf("read workspace manifest: %w", err)
	}
	return workspace, nil
}

func hash(data []byte) string { sum := sha256.Sum256(data); return fmt.Sprintf("%x", sum) }

func exclusiveWrite(path string, data []byte, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if os.IsExist(err) {
		existing, readErr := os.ReadFile(path)
		if readErr == nil && sha256.Sum256(existing) == sha256.Sum256(data) {
			return nil
		}
		return fmt.Errorf("immutable path already exists: %q", path)
	}
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return syncDir(filepath.Dir(path))
}

func removeAndSync(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return syncDir(filepath.Dir(path))
}

func syncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".tmp-")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	if err := file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, path); err != nil {
		return err
	}
	return syncDir(filepath.Dir(path))
}
