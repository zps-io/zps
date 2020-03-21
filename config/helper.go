package config

var ZshHelper = `###########
# zsh only, bash support eventually
###########

ZPS_IMAGE_DEFAULT=%s

zps_setup() {
    ZPS_SESSION=$(mktemp -t zps.sess.XXXXXX)
    export ZPS_SESSION

    precmd_functions+=(zps_reload)
    typeset -U precmd_functions

    if [[ -n "$ZPS_IMAGE" ]]; then
        return
    fi

    ZPS_PREV_PATH=$PATH
    ZPS_IMAGE=${ZPS_IMAGE_DEFAULT}

    PATH=${ZPS_IMAGE}/usr/bin:$ZPS_PREV_PATH

    export ZPS_PREV_PATH ZPS_IMAGE PATH
}

zps_reload() {
    local zps_update=$(cat "$ZPS_SESSION")

    if [[ -z "${zps_update}" ]]; then
        return
    fi

    : >! ${ZPS_SESSION}
    ZPS_IMAGE=${zps_update}
    PATH=${ZPS_IMAGE}/usr/bin:$ZPS_PREV_PATH

    typeset -U path
    export ZPS_IMAGE PATH
}

trap 'rm -f "$ZPS_SESSION"' EXIT

zps_setup
`
