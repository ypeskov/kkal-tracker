package email

// getSubject returns email subject based on language
func (s *Service) getSubject(language string) string {
	switch language {
	case "uk_UA":
		return "Активація облікового запису - Kkal Tracker"
	case "ru_UA":
		return "Активация учетной записи - Kkal Tracker"
	default:
		return "Account Activation - Kkal Tracker"
	}
}

// Email templates for activation emails in different languages
const activationTemplateEN = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Account Activation</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 8px;">
        <h1 style="color: #3b82f6; margin-bottom: 20px;">Welcome to {{.AppName}}!</h1>

        <p>Thank you for registering! To activate your account, please click the button below:</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ActivationURL}}"
               style="background-color: #3b82f6; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: bold;">
                Activate Account
            </a>
        </div>

        <p>Or copy and paste this link into your browser:</p>
        <p style="background-color: #fff; padding: 10px; border-radius: 4px; word-break: break-all;">
            <a href="{{.ActivationURL}}" style="color: #3b82f6;">{{.ActivationURL}}</a>
        </p>

        <p style="color: #6b7280; font-size: 14px; margin-top: 30px;">
            This activation link will expire in 24 hours.<br>
            If you didn't register for this account, please ignore this email.
        </p>
    </div>
</body>
</html>
`

const activationTemplateUK = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Активація облікового запису</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 8px;">
        <h1 style="color: #3b82f6; margin-bottom: 20px;">Ласкаво просимо до {{.AppName}}!</h1>

        <p>Дякуємо за реєстрацію! Щоб активувати свій обліковий запис, натисніть кнопку нижче:</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ActivationURL}}"
               style="background-color: #3b82f6; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: bold;">
                Активувати обліковий запис
            </a>
        </div>

        <p>Або скопіюйте та вставте це посилання у ваш браузер:</p>
        <p style="background-color: #fff; padding: 10px; border-radius: 4px; word-break: break-all;">
            <a href="{{.ActivationURL}}" style="color: #3b82f6;">{{.ActivationURL}}</a>
        </p>

        <p style="color: #6b7280; font-size: 14px; margin-top: 30px;">
            Це посилання для активації дійсне 24 години.<br>
            Якщо ви не реєструвалися, просто ігноруйте цей лист.
        </p>
    </div>
</body>
</html>
`

const activationTemplateRU = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Активация учетной записи</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 8px;">
        <h1 style="color: #3b82f6; margin-bottom: 20px;">Добро пожаловать в {{.AppName}}!</h1>

        <p>Спасибо за регистрацию! Чтобы активировать свою учетную запись, нажмите кнопку ниже:</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ActivationURL}}"
               style="background-color: #3b82f6; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: bold;">
                Активировать учетную запись
            </a>
        </div>

        <p>Или скопируйте и вставьте эту ссылку в ваш браузер:</p>
        <p style="background-color: #fff; padding: 10px; border-radius: 4px; word-break: break-all;">
            <a href="{{.ActivationURL}}" style="color: #3b82f6;">{{.ActivationURL}}</a>
        </p>

        <p style="color: #6b7280; font-size: 14px; margin-top: 30px;">
            Эта ссылка для активации действительна 24 часа.<br>
            Если вы не регистрировались, просто проигнорируйте это письмо.
        </p>
    </div>
</body>
</html>
`
