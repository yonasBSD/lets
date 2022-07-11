load test_helpers

setup() {
    load "${BATS_UTILS_PATH}/bats-support/load.bash"
    load "${BATS_UTILS_PATH}/bats-assert/load.bash"
    cd ./tests/zsh_completion
    cleanup
    cleanup_completion
}

create_completion() {
    lets completion --shell=zsh > _lets
}

@test "zsh_completion: should complete run command" {
    create_completion
    run ./completion_helper.sh "lets r"

    assert_success
    assert_output "run"
}

@test "zsh_completion: should complete run command options" {
    create_completion
    run ./completion_helper.sh "lets run --"

    assert_success
    assert_output <<EOF
--debug
--env
EOF
}

@test "zsh_completion: should complete run command options: --debug" {
    create_completion
    run ./completion_helper.sh "lets run --d"

    assert_success
    assert_output "--debug"
}

@test "zsh_completion: should complete run command options: --env" {
    create_completion
    run ./completion_helper.sh "lets run --e"

    assert_success
    assert_output "--env"
}

# TODO test lets own options copletions
# TODO test completions for bash - https://stackoverflow.com/questions/65386043/unit-testing-zsh-completion-script/69164362#69164362