package wailsapp

import (
	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/usecase"
)

type TransferBindings struct {
	app *App
	uc  *usecase.TransferUseCase
}

func NewTransferBindings(app *App, uc *usecase.TransferUseCase) *TransferBindings {
	return &TransferBindings{app: app, uc: uc}
}

func (b *TransferBindings) UploadFile(localPath, remotePath string) (string, error) {
	id, err := b.uc.Upload(b.app.Ctx(), localPath, remotePath)
	return string(id), err
}

func (b *TransferBindings) DownloadFile(remotePath, localPath string) (string, error) {
	id, err := b.uc.Download(b.app.Ctx(), remotePath, localPath)
	return string(id), err
}

func (b *TransferBindings) DownloadFolder(remoteRoot, localRoot string) (string, error) {
	id, err := b.uc.DownloadFolder(b.app.Ctx(), remoteRoot, localRoot)
	return string(id), err
}

func (b *TransferBindings) DownloadFolderAsZip(remoteRoot, localZip string) (string, error) {
	id, err := b.uc.DownloadFolderAsZip(b.app.Ctx(), remoteRoot, localZip)
	return string(id), err
}

func (b *TransferBindings) CancelTransfer(id string) error {
	return b.uc.Cancel(domain.TransferID(id))
}

func (b *TransferBindings) ListTransfers() []domain.TransferProgress {
	return b.uc.List()
}
