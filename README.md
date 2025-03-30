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
   **username** и **display name** укажите имя бота и запишите его в **файл dev.env -> MM_BOTNAME="vote-bot"** (например "vote-bot"). Поставьте **enabled** у **post:channel**.
8. Сгенерируется токен бота, который необходимо скопировать и вставить в **файл dev.env -> MM_TOKEN="YOUR_TOKEN"** (например "x3613cm7efyetksednieqpncpo").
9. Добавьте бота в вашу team: нажмите "+" рядом с названием team, выберите invite people и введите имя бота (например "@vote-bot").
10. Если нужно, создайте новые каналы: тот же "+" и опция "Create new channel".
11. Выберите, в каких каналах будет работать этот бот и впишите их названия в **файл dev.env -> MM_CHANNEL="test_channel,mychan"** (например "test_channel,mychan").
12. Добавьте бота в выбранные каналы: нажмите три точки рядом с названием канала и выберите add members, введите имя бота (например "@vote-bot").
13. Запустите команду **"make bot_db"** - развернутся контейнеры с самим ботом и базой данных tarantool. После запуска контейнеров в выбранных Вами чатах появится приветственное сообщение от бота. Все готово, можно им пользоваться.
14. Для вывода списка команд бота введите в чат канала **"@vote-bot help"**, вы должны увидеть:
```
for use bot enter folowing commands with format:

create voting - use minimum 3 lines:
  1.     @vote-bot new optional_exp_date(hh:mm:ss-dd:mm:yyyy)
  2.     voting name
  3.     option1 name
  n.     in next lines option2 name, option3 name, ...
vote:
  @vote-bot vote vote_id(number) var(number)
show voting with id:
  @vote-bot show vote_id(number)
show all:
  @vote-bot show_all
voting owner can close it:
  @vote-bot close vote_id(number)
voting owner can delete it:
  @vote-bot close vote_id(number)
```
