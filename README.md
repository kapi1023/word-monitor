
# word-monitor

Aplikacja CLI w Go do automatycznego monitorowania wolnych terminów egzaminów w serwisie info-car.pl (WORD).

## Funkcje

- 🔐 Logowanie do info-car z CSRF + token bearer
- 🧠 Pamięć ostatnio znanych terminów (dzień + godzina + ID praktyczne/teoretyczne)
- 📦 `state.json` z możliwością przechowywania wielu slotów egzaminacyjnych
- 🕓 Monitorowanie terminów co X sekund (configurable)
- 📢 Wysyłka powiadomień na Discord (webhook)
- 🌊Obsługa Dockera

## Instalacja

### Wymagania

- Go 1.22 lub nowszy

### Budowanie ze źródeł

```bash
git clone https://github.com/kapi1023/word-monitor.git
cd word-monitor
go build -o word-monitor ./cmd/monitor
```

## Uruchamianie z Dockerem
```bash
docker build -t word-monitor .
docker run -it --name word-monitor -p 2115:2115 word-monitor
```
## Licencja
Ten projekt jest objęty licencją MIT.
