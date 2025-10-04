package gui

import (
	"fmt"
	"net/url"

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

	// 创建密码列表，增加列宽度和操作按钮
	a.entryList = widget.NewList(
		func() int {
			return len(a.entries)
		},
		func() fyne.CanvasObject {
			// 创建可点击的标题按钮，设置为透明样式
			titleBtn := widget.NewButton("标题", func() {
				// 点击功能将在更新时设置
			})
			titleBtn.Resize(fyne.NewSize(150, 30)) // 减少标题宽度
			titleBtn.Importance = widget.LowImportance // 设置为低重要性，减少按钮样式
			
			// 创建可点击的用户名按钮，设置为透明样式
			usernameBtn := widget.NewButton("用户名", func() {
				// 点击功能将在更新时设置
			})
			usernameBtn.Resize(fyne.NewSize(100, 30)) // 减少用户名宽度
			usernameBtn.Importance = widget.LowImportance // 设置为低重要性，减少按钮样式
			
			// 创建URL容器，使用无布局容器但设置合适的尺寸
			urlContainer := container.NewWithoutLayout()
			urlContainer.Resize(fyne.NewSize(280, 30)) // 给URL更多空间
			
			// 创建编辑按钮
			editBtn := widget.NewButton("编辑", func() {
				// 编辑功能将在更新时设置
			})
			editBtn.Resize(fyne.NewSize(60, 30))
			
			// 创建删除按钮
			deleteBtn := widget.NewButton("删除", func() {
				// 删除功能将在更新时设置
			})
			deleteBtn.Resize(fyne.NewSize(60, 30))
			
			// 创建复制按钮
			copyBtn := widget.NewButton("复制", func() {
				// 复制功能将在更新时设置
			})
			copyBtn.Resize(fyne.NewSize(60, 30))
			
			// 创建按钮容器
			buttonContainer := container.NewHBox(editBtn, deleteBtn, copyBtn)
			
			// 使用更简单的布局结构，避免事件冲突
			infoContainer := container.NewHBox(
				titleBtn,
				usernameBtn,
				urlContainer,
			)
			
			return container.NewBorder(
				nil, // 顶部
				nil, // 底部
				nil, // 左侧
				buttonContainer, // 右侧：操作按钮
				infoContainer, // 中心：信息标签
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(a.entries) {
				return
			}
			entry := a.entries[id]
			borderContainer := obj.(*fyne.Container)
			
			// 获取中心的信息容器
			infoContainer := borderContainer.Objects[0].(*fyne.Container)
			titleBtn := infoContainer.Objects[0].(*widget.Button)
			usernameBtn := infoContainer.Objects[1].(*widget.Button)
			urlContainer := infoContainer.Objects[2].(*fyne.Container)
			
			// 设置标题和用户名文本
			titleBtn.SetText(entry.Title)
			usernameBtn.SetText(entry.Username)
			
			// 设置点击事件，点击标题或用户名都显示详情
			titleBtn.OnTapped = func() {
				a.showEntryDetails(entry)
			}
			usernameBtn.OnTapped = func() {
				a.showEntryDetails(entry)
			}
			
			// 更新URL容器内容，确保URL链接能正常工作
			urlContainer.RemoveAll()
			if entry.URL != "" {
				urlWidget := a.createURLWidget(entry.URL)
				// 对于无布局容器，需要手动设置位置和大小
				urlWidget.Resize(fyne.NewSize(280, 30))
				urlWidget.Move(fyne.NewPos(0, 0))
				urlContainer.Add(urlWidget)
			} else {
				noUrlLabel := widget.NewLabel("无")
				noUrlLabel.Resize(fyne.NewSize(280, 30))
				noUrlLabel.Move(fyne.NewPos(0, 0))
				urlContainer.Add(noUrlLabel)
			}
			
			// 获取右侧的按钮容器
			buttonContainer := borderContainer.Objects[1].(*fyne.Container)
			editBtn := buttonContainer.Objects[0].(*widget.Button)
			deleteBtn := buttonContainer.Objects[1].(*widget.Button)
			copyBtn := buttonContainer.Objects[2].(*widget.Button)
			
			// 设置编辑按钮功能
			editBtn.OnTapped = func() {
				a.showEntryDialog(entry)
			}
			
			// 设置删除按钮功能
			deleteBtn.OnTapped = func() {
				a.showCustomConfirmDialog("确认删除", "确定要删除这个密码条目吗？", func(confirmed bool) {
					if confirmed {
						if err := a.db.DeletePasswordEntry(entry.ID); err != nil {
							dialog.ShowError(err, a.window)
							return
						}
						a.loadEntries()
					}
				})
			}
			
			// 设置复制按钮功能
			copyBtn.OnTapped = func() {
				// 格式化复制内容：账号和密码换行显示
				// entry.Password 已经是解密后的明文密码，无需再次解密
				copyContent := fmt.Sprintf("账号: %s\n密码: %s", entry.Username, entry.Password)
				
				// 复制到剪切板
				a.window.Clipboard().SetContent(copyContent)
				
				// 显示复制成功提示
				dialog.ShowInformation("复制成功", "账号和密码已复制到剪切板", a.window)
			}
		},
	)

	// 移除双击查看详情的OnSelected事件，避免与URL点击冲突
	// 现在只能通过点击标题或用户名来查看详情
	// a.entryList.OnSelected = func(id widget.ListItemID) {
	// 	if id >= len(a.entries) {
	// 		return
	// 	}
	// 	a.showEntryDetails(a.entries[id])
	// 	// 显示详情后立即取消选中，确保下次点击同一项目时能再次触发
	// 	a.entryList.UnselectAll()
	// }

	// 创建工具栏按钮
	addButton := widget.NewButton("添加密码", func() {
		a.showAddEntryDialog()
	})
	addButton.Resize(fyne.NewSize(100, 35))

	// 创建工具栏容器
	toolbar := container.NewHBox(
		addButton,
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
	titleLabel.Wrapping = fyne.TextWrapWord
	
	usernameLabel := widget.NewLabel(entry.Username)
	usernameLabel.Wrapping = fyne.TextWrapWord
	
	passwordLabel := widget.NewLabel("••••••••")
	passwordLabel.Wrapping = fyne.TextWrapWord
	
	categoryLabel := widget.NewLabel(entry.Category)
	categoryLabel.Wrapping = fyne.TextWrapWord
	
	notesLabel := widget.NewLabel(entry.Notes)
	notesLabel.Wrapping = fyne.TextWrapWord

	showPasswordBtn := widget.NewButton("显示密码", func() {
		if passwordLabel.Text == "••••••••" {
			passwordLabel.SetText(entry.Password)
		} else {
			passwordLabel.SetText("••••••••")
		}
	})

	// 创建关闭按钮
	closeBtn := widget.NewButton("关闭", func() {
		// 关闭功能将在对话框创建后设置
	})

	// 创建顶部容器，关闭按钮在右上角
	topContainer := container.NewBorder(
		nil, // 顶部
		nil, // 底部
		nil, // 左侧
		closeBtn, // 右侧：关闭按钮
		widget.NewLabel(""), // 中心：空白占位
	)

	// 创建URL容器，确保URL能正确显示
	urlWidget := a.createURLWidget(entry.URL)
	urlContainer := container.NewGridWithColumns(2,
		widget.NewLabel("网址:"), urlWidget,
	)

	// 创建详情内容，使用更好的布局
	detailsContent := container.NewVBox(
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("标题:"), titleLabel,
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("用户名:"), usernameLabel,
		),
		widget.NewSeparator(),
		container.NewBorder(
			nil, nil, widget.NewLabel("密码:"), showPasswordBtn,
			passwordLabel,
		),
		widget.NewSeparator(),
		urlContainer,
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("分类:"), categoryLabel,
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("备注:"), notesLabel,
		),
		widget.NewSeparator(),
	)

	// 创建完整内容容器，移除滚动条
	content := container.NewBorder(
		topContainer, // 顶部：关闭按钮在右边
		nil, // 底部
		nil, // 左侧
		nil, // 右侧
		container.NewPadded(detailsContent), // 中心：添加内边距的详情内容
	)

	// 创建详情对话框，设置合适的大小
	detailsDialog := dialog.NewCustomWithoutButtons("密码详情", content, a.window)
	detailsDialog.Resize(fyne.NewSize(600, 450))

	// 设置关闭按钮功能
	closeBtn.OnTapped = func() {
		detailsDialog.Hide()
	}

	detailsDialog.Show()
}

// createURLWidget 创建可点击的URL组件
func (a *App) createURLWidget(urlStr string) fyne.CanvasObject {
	if urlStr == "" {
		return widget.NewLabel("无")
	}
	
	// 验证URL格式
	parsedURL, err := url.Parse(urlStr)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		// 如果不是有效的HTTP/HTTPS URL，显示为普通文本，不换行
		label := widget.NewLabel(urlStr)
		label.Wrapping = fyne.TextWrapOff // 关闭换行
		return label
	}
	
	// 创建可点击的超链接，确保URL能在浏览器中打开
	hyperlink := widget.NewHyperlink(urlStr, parsedURL)
	// 设置超链接样式，关闭换行，横向显示
	hyperlink.Wrapping = fyne.TextWrapOff // 关闭换行
	hyperlink.Truncation = fyne.TextTruncateOff // 关闭截断，显示完整URL
	return hyperlink
}

// showCustomConfirmDialog 显示自定义确认对话框，"是"在左边，"否"在右边
func (a *App) showCustomConfirmDialog(title, message string, callback func(bool)) {
	// 创建消息标签，居中对齐
	messageLabel := widget.NewLabel(message)
	messageLabel.Alignment = fyne.TextAlignCenter
	
	// 创建按钮，设置更大的尺寸
	yesBtn := widget.NewButton("是", func() {
		callback(true)
	})
	yesBtn.Resize(fyne.NewSize(80, 40))
	
	noBtn := widget.NewButton("否", func() {
		callback(false)
	})
	noBtn.Resize(fyne.NewSize(80, 40))
	
	// 创建按钮容器，使用 HBox 并添加间距，然后居中
	buttonContainer := container.NewHBox(
		yesBtn,
		widget.NewLabel("   "), // 添加间距
		noBtn,
	)
	
	// 创建内容容器，使用 VBox 并居中对齐
	content := container.NewVBox(
		widget.NewLabel(""), // 顶部间距
		messageLabel,
		widget.NewLabel(""), // 中间间距
		container.NewCenter(buttonContainer), // 按钮容器居中
		widget.NewLabel(""), // 底部间距
	)
	
	// 创建对话框
	confirmDialog := dialog.NewCustomWithoutButtons(title, content, a.window)
	confirmDialog.Resize(fyne.NewSize(300, 150))
	
	// 设置按钮功能
	yesBtn.OnTapped = func() {
		confirmDialog.Hide()
		callback(true)
	}
	
	noBtn.OnTapped = func() {
		confirmDialog.Hide()
		callback(false)
	}
	
	confirmDialog.Show()
}