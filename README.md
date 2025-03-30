# mattermost-bot

Чат бот для корпоративного мессенджера mattermost.

## Запуск
Для запуска необходимо:
1. Скопировать проект к себе на локальную машину.
2. Убедитесь, что Docker Engine запущен, порты 8065 и 3301 свободны.
3. Запустите команду **"make mattermost"** - развернется контейнер с месенджером mattermost (localhost:8065).
4. Перейдите по localhost:8065 и зарегестрируйтесь (эти данные далее не понадобятся).
5. Задайте имя организации и запишите его сразу в **файл dev.env -> MM_TEAM="myteam"** (например "myteam").
6. Пропускайте/подтверждайте следующие настройки и после того как попадете на главную страницу, переходите по http://localhost:8065/admin_console/integrations/bot_accounts
  значение **Enable Bot Account Creation:** укажите **True**, сохраните.
7. Далее переходите по http://localhost:8065/myteam/integrations/bots и создайте нового бота:
   **username** и **display name** укажите имя бота и запишите его в **файл dev.env -> MM_TEAM="myteam"** (например "vote-bot").
