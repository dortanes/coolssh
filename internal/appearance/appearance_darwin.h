#ifndef COOLSSH_APPEARANCE_H
#define COOLSSH_APPEARANCE_H

typedef enum {
    coolssh_appearance_system,
    coolssh_appearance_light,
    coolssh_appearance_dark,
} coolssh_appearance;

void coolssh_appearance_set(coolssh_appearance appearance);

#endif
