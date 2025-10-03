package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"hank.com/password_tool/crypto"
	"hank.com/password_tool/database"
	"hank.com/password_tool/models"
)

type App struct {
	fyneApp    fyne.App
	window     fyne.Window
	db         *database.DB
	entryList  *widget.List
	entries    []*models.PasswordEntry
	categories []*models.Category
}

// NewApp 创建新的应用实例
func NewApp() *App {
	fyneApp := app.New()
	fyneApp.SetIcon(nil) // 可以设置应用图标

	window := fyneApp.NewWindow("密码管理器")
	window.Resize(fyne.NewSize(1000, 700)) // 增加窗口尺寸，提供更好的用户体验
	window.CenterOnScreen()

	return &App{
		fyneApp: fyneApp,
		window:  window,
	}
}

// Run 运行应用
func (a *App) Run() {
	var err error
	a.db, err = database.NewDB()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	defer a.db.Close()

	// 检查是否已设置主密码
	hasMasterPassword, err := a.db.HasMasterPassword()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	if !hasMasterPassword {
		a.showSetMasterPasswordDialog()
	} else {
		a.showLoginDialog()
	}

	a.window.ShowAndRun()
}

// showSetMasterPasswordDialog 显示设置主密码对话框
func (a *App) showSetMasterPasswordDialog() {
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.Resize(fyne.NewSize(300, 40))
	
	confirmEntry := widget.NewPasswordEntry()
	confirmEntry.Resize(fyne.NewSize(300, 40))
	
	// 设置主密码处理函数
	setPasswordFunc := func() {
		password := passwordEntry.Text
		confirm := confirmEntry.Text

		if password == "" {
			dialog.ShowError(fmt.Errorf("密码不能为空"), a.window)
			return
		}

		if password != confirm {
			dialog.ShowError(fmt.Errorf("两次输入的密码不一致"), a.window)
			return
		}

		if err := a.db.SetMasterPassword(password); err != nil {
			dialog.ShowError(err, a.window)
			return
		}

		// 设置主密钥
		salt, err := a.db.GetMasterPasswordSalt()
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		key := crypto.DeriveKey(password, salt)
		a.db.SetMasterKey(key)

		a.showMainWindow()
	}

	// 添加回车键监听
	passwordEntry.OnSubmitted = func(text string) {
		confirmEntry.FocusGained()
	}
	
	confirmEntry.OnSubmitted = func(text string) {
		setPasswordFunc()
	}

	// 创建确定按钮
	confirmButton := widget.NewButton("确定", func() {
		setPasswordFunc()
	})
	confirmButton.Resize(fyne.NewSize(100, 35))

	// 创建标签
	passwordLabel := widget.NewLabel("主密码:")
	confirmLabel := widget.NewLabel("确认密码:")
	
	// 添加适当的间距
	spacer := widget.NewLabel("")
	spacer.Resize(fyne.NewSize(1, 15))
	
	content := container.NewVBox(
		spacer,
		passwordLabel,
		passwordEntry,
		spacer,
		confirmLabel,
		confirmEntry,
		spacer,
		container.NewCenter(confirmButton),
		spacer,
	)
	
	// 添加内边距
	paddedContent := container.NewPadded(content)
	
	// 设置主窗口标题和内容
	a.window.SetTitle("设置主密码")
	a.window.SetContent(paddedContent)
	a.window.Resize(fyne.NewSize(450, 250))
	a.window.CenterOnScreen()
}

// showLoginDialog 显示登录对话框
func (a *App) showLoginDialog() {
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.Resize(fyne.NewSize(300, 40))
	
	// 登录处理函数
	loginFunc := func() {
		password := passwordEntry.Text

		valid, err := a.db.VerifyMasterPassword(password)
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}

		if !valid {
			dialog.ShowError(fmt.Errorf("密码错误"), a.window)
			return
		}

		// 设置主密钥
		salt, err := a.db.GetMasterPasswordSalt()
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		key := crypto.DeriveKey(password, salt)
		a.db.SetMasterKey(key)

		a.showMainWindow()
	}

	// 添加回车键监听
	passwordEntry.OnSubmitted = func(text string) {
		loginFunc()
	}

	// 创建登录按钮
	loginButton := widget.NewButton("登录", func() {
		loginFunc()
	})
	loginButton.Resize(fyne.NewSize(100, 35))

	// 创建简单的标签和输入框布局
	label := widget.NewLabel("主密码:")
	
	// 添加适当的间距
	spacer := widget.NewLabel("")
	spacer.Resize(fyne.NewSize(1, 15))
	
	content := container.NewVBox(
		spacer,
		label,
		passwordEntry,
		spacer,
		container.NewCenter(loginButton),
		spacer,
	)
	
	// 添加内边距
	paddedContent := container.NewPadded(content)
	
	// 设置主窗口标题和内容
	a.window.SetTitle("输入主密码")
	a.window.SetContent(paddedContent)
	a.window.Resize(fyne.NewSize(380, 160))
	a.window.CenterOnScreen()
}

// showMainWindow 显示主窗口
func (a *App) showMainWindow() {
	a.loadEntries()

	// 创建密码列表，增加列宽度
	a.entryList = widget.NewList(
		func() int {
			return len(a.entries)
		},
		func() fyne.CanvasObject {
			titleLabel := widget.NewLabel("标题")
			titleLabel.Resize(fyne.NewSize(200, 30))
			
			usernameLabel := widget.NewLabel("用户名")
			usernameLabel.Resize(fyne.NewSize(150, 30))
			
			urlLabel := widget.NewLabel("网址")
			urlLabel.Resize(fyne.NewSize(200, 30))
			
			return container.NewHBox(
				titleLabel,
				usernameLabel,
				urlLabel,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(a.entries) {
				return
			}
			entry := a.entries[id]
			hbox := obj.(*fyne.Container)
			hbox.Objects[0].(*widget.Label).SetText(entry.Title)
			hbox.Objects[1].(*widget.Label).SetText(entry.Username)
			hbox.Objects[2].(*widget.Label).SetText(entry.URL)
		},
	)

	// 双击查看详情
	a.entryList.OnSelected = func(id widget.ListItemID) {
		if id >= len(a.entries) {
			return
		}
		a.showEntryDetails(a.entries[id])
	}

	// 创建工具栏按钮
	addButton := widget.NewButton("添加密码", func() {
		a.showAddEntryDialog()
	})
	addButton.Resize(fyne.NewSize(100, 35))
	
	refreshButton := widget.NewButton("刷新", func() {
		a.loadEntries()
	})
	refreshButton.Resize(fyne.NewSize(80, 35))

	// 创建工具栏容器
	toolbar := container.NewHBox(
		addButton,
		refreshButton,
	)

	// 创建搜索框，增加高度
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("搜索密码条目...")
	searchEntry.Resize(fyne.NewSize(0, 35)) // 宽度自适应，高度35
	searchEntry.OnChanged = func(text string) {
		a.filterEntries(text)
	}

	// 创建顶部容器，增加间距
	topContainer := container.NewVBox(
		container.NewPadded(toolbar),
		container.NewPadded(searchEntry),
	)

	// 布局
	content := container.NewBorder(
		topContainer,
		nil,
		nil,
		nil,
		container.NewPadded(a.entryList),
	)

	// 设置主窗口标题和内容
	a.window.SetTitle("密码管理器")
	a.window.SetContent(content)
	a.window.Resize(fyne.NewSize(800, 600))
	a.window.CenterOnScreen()
}

// loadEntries 加载密码条目
func (a *App) loadEntries() {
	entries, err := a.db.GetPasswordEntries()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.entries = entries
	if a.entryList != nil {
		a.entryList.Refresh()
	}
}

// filterEntries 过滤密码条目
func (a *App) filterEntries(searchText string) {
	if searchText == "" {
		a.loadEntries()
		return
	}

	allEntries, err := a.db.GetPasswordEntries()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	var filtered []*models.PasswordEntry
	for _, entry := range allEntries {
		if contains(entry.Title, searchText) ||
			contains(entry.Username, searchText) ||
			contains(entry.URL, searchText) ||
			contains(entry.Category, searchText) {
			filtered = append(filtered, entry)
		}
	}

	a.entries = filtered
	a.entryList.Refresh()
}

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(substr) == 0 || 
		    (len(s) > 0 && len(substr) > 0 && s[0] == substr[0]))
}

// showAddEntryDialog 显示添加条目对话框
func (a *App) showAddEntryDialog() {
	a.showEntryDialog(nil)
}

// showEntryDialog 显示条目编辑对话框
func (a *App) showEntryDialog(entry *models.PasswordEntry) {
	// 创建输入框并设置尺寸
	titleEntry := widget.NewEntry()
	titleEntry.Resize(fyne.NewSize(350, 35))
	
	usernameEntry := widget.NewEntry()
	usernameEntry.Resize(fyne.NewSize(350, 35))
	
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.Resize(fyne.NewSize(350, 35))
	
	urlEntry := widget.NewEntry()
	urlEntry.Resize(fyne.NewSize(350, 35))
	
	categoryEntry := widget.NewEntry()
	categoryEntry.Resize(fyne.NewSize(350, 35))
	
	notesEntry := widget.NewMultiLineEntry()
	notesEntry.Resize(fyne.NewSize(350, 80))

	// 如果是编辑模式，填充现有数据
	if entry != nil {
		titleEntry.SetText(entry.Title)
		usernameEntry.SetText(entry.Username)
		passwordEntry.SetText(entry.Password)
		urlEntry.SetText(entry.URL)
		notesEntry.SetText(entry.Notes)
		categoryEntry.SetText(entry.Category)
	}

	// 创建标签，设置固定宽度以确保对齐
	titleLabel := widget.NewLabel("标题:")
	usernameLabel := widget.NewLabel("用户名:")
	passwordLabel := widget.NewLabel("密码:")
	urlLabel := widget.NewLabel("网址:")
	categoryLabel := widget.NewLabel("分类:")
	notesLabel := widget.NewLabel("备注:")
	
	// 使用网格布局创建表单，2列布局：标签列和输入框列
	formContent := container.NewGridWithColumns(2,
		titleLabel, titleEntry,
		usernameLabel, usernameEntry,
		passwordLabel, passwordEntry,
		urlLabel, urlEntry,
		categoryLabel, categoryEntry,
		notesLabel, notesEntry,
	)

	// 添加垂直间距和内边距的容器，不使用Card组件避免额外按钮
	paddedContent := container.NewPadded(formContent)

	// 确定对话框标题
	title := "添加密码"
	if entry != nil {
		title = "编辑密码"
	}

	// 创建关闭按钮
	closeButton := widget.NewButton("关闭", func() {
		// 关闭对话框的逻辑将在对话框创建后设置
	})
	
	// 创建保存按钮
	saveButton := widget.NewButton("保存", func() {
		if titleEntry.Text == "" || passwordEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("标题和密码不能为空"), a.window)
			return
		}

		newEntry := &models.PasswordEntry{
			Title:    titleEntry.Text,
			Username: usernameEntry.Text,
			Password: passwordEntry.Text,
			URL:      urlEntry.Text,
			Notes:    notesEntry.Text,
			Category: categoryEntry.Text,
		}

		var err error
		if entry == nil {
			err = a.db.AddPasswordEntry(newEntry)
		} else {
			newEntry.ID = entry.ID
			err = a.db.UpdatePasswordEntry(newEntry)
		}

		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}

		a.loadEntries()
		// 关闭对话框的逻辑将在对话框创建后设置
	})
	
	// 创建顶部容器，关闭按钮在最右边
	topContainer := container.NewBorder(
		nil, // 顶部
		nil, // 底部
		nil, // 左侧
		closeButton, // 右侧：关闭按钮
		widget.NewLabel(""), // 中心：空白占位
	)
	
	// 创建底部容器，保存按钮居中
	bottomContainer := container.NewCenter(saveButton)
	
	// 创建完整的内容容器
	fullContent := container.NewBorder(
		topContainer, // 顶部：关闭按钮在右边
		bottomContainer, // 底部：保存按钮居中
		nil, // 左侧
		nil, // 右侧
		paddedContent, // 中心：表单内容
	)

	// 创建自定义对话框，使用 NewCustomWithoutButtons 避免底部默认按钮
	d := dialog.NewCustomWithoutButtons(title, fullContent, a.window)
	
	// 设置关闭按钮和保存按钮的关闭对话框功能
	closeButton.OnTapped = func() {
		d.Hide()
	}
	
	// 更新保存按钮的关闭对话框功能
	saveButton.OnTapped = func() {
		if titleEntry.Text == "" || passwordEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("标题和密码不能为空"), a.window)
			return
		}

		newEntry := &models.PasswordEntry{
			Title:    titleEntry.Text,
			Username: usernameEntry.Text,
			Password: passwordEntry.Text,
			URL:      urlEntry.Text,
			Notes:    notesEntry.Text,
			Category: categoryEntry.Text,
		}

		var err error
		if entry == nil {
			err = a.db.AddPasswordEntry(newEntry)
		} else {
			newEntry.ID = entry.ID
			err = a.db.UpdatePasswordEntry(newEntry)
		}

		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}

		a.loadEntries()
		d.Hide()
	}
	
	d.Resize(fyne.NewSize(500, 500))
	d.Show()
}

// showEntryDetails 显示条目详情
func (a *App) showEntryDetails(entry *models.PasswordEntry) {
	titleLabel := widget.NewLabel(entry.Title)
	usernameLabel := widget.NewLabel(entry.Username)
	passwordLabel := widget.NewLabel("••••••••")
	urlLabel := widget.NewLabel(entry.URL)
	categoryLabel := widget.NewLabel(entry.Category)
	notesLabel := widget.NewLabel(entry.Notes)

	showPasswordBtn := widget.NewButton("显示密码", func() {
		if passwordLabel.Text == "••••••••" {
			passwordLabel.SetText(entry.Password)
		} else {
			passwordLabel.SetText("••••••••")
		}
	})

	// 创建详情对话框
	content := container.NewVBox(
		widget.NewCard("", "", container.NewVBox(
			container.NewHBox(widget.NewLabel("标题:"), titleLabel),
			container.NewHBox(widget.NewLabel("用户名:"), usernameLabel),
			container.NewHBox(widget.NewLabel("密码:"), passwordLabel, showPasswordBtn),
			container.NewHBox(widget.NewLabel("网址:"), urlLabel),
			container.NewHBox(widget.NewLabel("分类:"), categoryLabel),
			container.NewHBox(widget.NewLabel("备注:"), notesLabel),
		)),
	)

	// 创建详情对话框
	detailsDialog := dialog.NewCustom("密码详情", "关闭", content, a.window)

	// 创建编辑按钮，在编辑完成后关闭详情对话框
	editBtn := widget.NewButton("编辑", func() {
		detailsDialog.Hide() // 先关闭详情对话框
		a.showEntryDialog(entry)
	})

	deleteBtn := widget.NewButton("删除", func() {
		dialog.ShowConfirm("确认删除", "确定要删除这个密码条目吗？", func(confirmed bool) {
			if confirmed {
				if err := a.db.DeletePasswordEntry(entry.ID); err != nil {
					dialog.ShowError(err, a.window)
					return
				}
				a.loadEntries()
				detailsDialog.Hide() // 删除后关闭详情对话框
			}
		}, a.window)
	})

	// 添加按钮到内容中
	content.Add(container.NewHBox(editBtn, deleteBtn))

	detailsDialog.Show()
}