package icons

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// GetAllIcons returns all available theme icons
func GetAllIcons() []fyne.Resource {
	return []fyne.Resource{
		theme.FyneLogo(),
		theme.HomeIcon(),
		theme.InfoIcon(),
		theme.MailAttachmentIcon(),
		theme.MediaPlayIcon(),
		theme.MediaRecordIcon(),
		theme.MenuIcon(),
		theme.MoveDownIcon(),
		theme.NavigateBackIcon(),
		theme.NavigateNextIcon(),
		theme.RadioButtonCheckedIcon(),
		theme.SearchIcon(),
		theme.SettingsIcon(),
		theme.StorageIcon(),
		theme.ViewRefreshIcon(),
		theme.WarningIcon(),
		theme.VisibilityIcon(),
		theme.UploadIcon(),
		theme.VolumeUpIcon(),
		theme.VolumeMuteIcon(),
		theme.ViewFullScreenIcon(),
		theme.ViewRestoreIcon(),
		theme.VisibilityOffIcon(),
		theme.ZoomInIcon(),
		theme.ZoomOutIcon(),
		theme.ZoomFitIcon(),
		theme.MailForwardIcon(),
		theme.MailReplyIcon(),
		theme.MailReplyAllIcon(),
		theme.MailSendIcon(),
	}
}
