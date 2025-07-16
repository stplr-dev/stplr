# This file was originally part of the project "LURE - Linux User REpository", created by Elara Musayelyan.
# It has been modified as part of "ALR - Any Linux Repository" by the ALR Authors.
#
# ALR - Any Linux Repository
# Copyright (C) 2025 The ALR Authors
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

info() {
  echo $'\x1b[32m[ИНФО]\x1b[0m' $@
}

warn() {
  echo $'\x1b[31m[ВНИМАНИЕ]\x1b[0m' $@
}

error() {
  echo $'\x1b[31;1m[ОШИБКА]\x1b[0m' $@
  exit 1
}

installPkg() {
  rootCmd=""
  if command -v doas &>/dev/null; then
    rootCmd="doas"
  elif command -v sudo &>/dev/null; then
    rootCmd="sudo"
  else
    warn "Не обнаружена команда повышения привилегий (например, sudo, doas)"
  fi

  case $1 in
  pacman) $rootCmd pacman --noconfirm -U "${@:2}" ;;
  apk) $rootCmd apk add --allow-untrusted "${@:2}" ;;
  zypper) $rootCmd zypper --no-gpg-checks install "${@:2}" ;;
  *) $rootCmd "$1" install -y "${@:2}" ;;
  esac
}

if ! command -v curl &>/dev/null; then
  error "Этот скрипт требует команду curl. Пожалуйста, установите её и запустите снова."
fi

pkgFormat=""
pkgMgr=""
if command -v pacman &>/dev/null; then
  info "Обнаружен pacman"
  pkgFormat="pkg.tar.zst"
  pkgMgr="pacman"
elif command -v apt &>/dev/null; then
  info "Обнаружен apt"
  pkgFormat="deb"
  pkgMgr="apt"
elif command -v dnf &>/dev/null; then
  info "Обнаружен dnf"
  pkgFormat="rpm"
  pkgMgr="dnf"
elif command -v yum &>/dev/null; then
  info "Обнаружен yum"
  pkgFormat="rpm"
  pkgMgr="yum"
elif command -v zypper &>/dev/null; then
  info "Обнаружен zypper"
  pkgFormat="rpm"
  pkgMgr="zypper"
elif command -v apk &>/dev/null; then
  info "Обнаружен apk"
  pkgFormat="apk"
  pkgMgr="apk"
elif command -v apt-get &>/dev/null; then
  info "Обнаружен apt-get"
  pkgFormat="rpm"
  pkgMgr="apt-get"
else
  warn "Не обнаружен поддерживаемый менеджер пакетов!"
  noPkgMgr=true
fi

if [ -z "$noPkgMgr" ]; then
  info "Получение списка файлов с https://go.stplr.dev/stplr/releases"

  # Изменено URL и регулярное выражение для списка файлов
  pageContent=$(curl -s https://go.stplr.dev/stplr/releases)

  # Извлечение списка файлов из HTML
  fileList=$(echo "$pageContent" | grep -oP '(?<=href=").*?(?=")' | grep -E 'alr-bin.*\.(pkg.tar.zst|rpm|deb)')

  echo "Полученный список файлов:"
  echo "$fileList"

  if [ "$pkgMgr" == "pacman" ]; then
      latestFile=$(echo "$fileList" | grep -E 'alr-bin-.*\.pkg\.tar\.zst' | sort -V | tail -n 1)
  elif [ "$pkgMgr" == "apt" ]; then
      latestFile=$(echo "$fileList" | grep -E 'alr-bin-.*\.amd64\.deb' | sort -V | tail -n 1)
  elif [[ "$pkgMgr" == "dnf" || "$pkgMgr" == "yum" || "$pkgMgr" == "zypper" ]]; then
      latestFile=$(printf "%s\n" "${fileList[@]}" | grep -E 'alr-bin-.*\.x86_64\.rpm' | grep -v 'alt[0-9]*' | sort -V | tail -n 1)
  elif [ "$pkgMgr" == "apt-get" ]; then
      latestFile=$(echo "$fileList" | grep -E 'alr-bin-.*-alt[0-9]+\.x86_64\.rpm' | sort -V | tail -n 1)
  else
      error "Не поддерживаемый менеджер пакетов для автоматической установки"
  fi

  if [ -z "$latestFile" ]; then
      error "Не удалось найти соответствующий пакет для $pkgMgr"
  fi

  info "Найдена последняя версия ALR: $latestFile"

  fname="$(mktemp -u -p /tmp "alr.XXXXXXXXXX").${pkgFormat}"

  info "Загрузка пакета ALR"
  curl -o $fname -L "$latestFile"

  if [ ! -f "$fname" ]; then
      error "Ошибка загрузки пакета ALR"
  fi

  info "Установка пакета ALR"
  installPkg "$pkgMgr" "$fname"

  info "Очистка"
  rm "$fname"

  info "Готово!"
else
  echo "Не найден поддерживаемый менеджер пакетов. О_о"
fi
