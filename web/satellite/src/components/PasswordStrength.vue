// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

<template>
    <v-card
        class="positioning mt-n2"
        width="100%"
        position="absolute"
        elevation="12"
        variant="elevated"
        :theme="oppositeTheme"
    >
        <v-card-item class="mt-1">
            <v-card-title>Password Security Level</v-card-title>
            <template #append>
                <v-btn
                    icon="$close"
                    variant="text"
                    size="x-small"
                    color="default"
                />
            </template>
        </v-card-item>
        <v-card-item>
            <p class="text-body-2 font-weight-bold mb-2" :style="strengthLabelColor">{{ passwordStrength }}</p>
            <v-progress-linear :model-value="barWidth" :color="passwordStrengthColor" rounded="lg" />
        </v-card-item>
        <v-card-item>
            <p class="text-body-2">Your password should contain:</p>
            <v-checkbox
                tabindex="-1"
                class="no-pointer-events text-body-2"
                :model-value="isPasswordLengthAcceptable"
                color="success"
                density="compact"
                hide-details
            >
                <template #label>
                    <p class="text-body-2">Between {{ passMinLength }} and {{ passMaxLength }} Latin characters</p>
                </template>
            </v-checkbox>
            <p class="text-body-2 mt-2">Its nice to have:</p>
            <v-checkbox
                tabindex="-1"
                class="no-pointer-events text-body-2"
                :model-value="hasLowerAndUpperCaseLetters"
                color="success"
                density="compact"
                hide-details
            >
                <template #label>
                    <p class="text-body-2">Upper and lowercase letters</p>
                </template>
            </v-checkbox>
            <v-checkbox
                tabindex="-1"
                class="no-pointer-events text-body-2 mt-n2"
                :model-value="hasSpecialCharacter"
                color="success"
                density="compact"
                hide-details
            >
                <template #label>
                    <p class="text-body-2">At least one special character</p>
                </template>
            </v-checkbox>
            <p class="text-caption mt-2">
                Avoid using a password that you use on other websites or that might be easily guessed by someone else.
            </p>
        </v-card-item>
    </v-card>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useTheme } from 'vuetify';
import {
    VCard,
    VCardItem,
    VProgressLinear,
    VCheckbox,
    VCardTitle,
    VBtn,
} from 'vuetify/components';

import { useConfigStore } from '@/store/modules/configStore';

const configStore = useConfigStore();

const theme = useTheme();

const oppositeTheme = computed(() => {
    return theme.current.value.dark ? 'light' : 'dark';
});

const PASSWORD_STRENGTH = {
    veryStrong: 'Very Strong',
    strong: 'Strong',
    good: 'Good',
    weak: 'Weak',
};

const PASSWORD_STRENGTH_COLORS = {
    [PASSWORD_STRENGTH.good]: '#ffa500',
    [PASSWORD_STRENGTH.strong]: '#00AC26',
    [PASSWORD_STRENGTH.veryStrong]: '#00AC26',
    default: '#ff0000',
};

const BAR_WIDTH = {
    [PASSWORD_STRENGTH.weak]: '25',
    [PASSWORD_STRENGTH.good]: '50',
    [PASSWORD_STRENGTH.strong]: '75',
    [PASSWORD_STRENGTH.veryStrong]: '100',
    default: '0',
};

const props = defineProps<{
    password: string;
    email?: string;
}>();

/**
 * Returns the maximum password length from the store.
 */
const passMaxLength = computed((): number => {
    return configStore.state.config.passwordMaximumLength;
});

/**
 * Returns the minimum password length from the store.
 */
const passMinLength = computed((): number => {
    return configStore.state.config.passwordMinimumLength;
});

const isPasswordLengthAcceptable = computed((): boolean => {
    return props.password.length <= passMaxLength.value
        && props.password.length >= passMinLength.value;
});

/**
 * Returns password strength label depends on score.
 */
const passwordStrength = computed((): string => {
    if (props.password.length < passMinLength.value) {
        return `Use ${passMinLength.value} or more characters`;
    }

    if (props.password.length > passMaxLength.value) {
        return `Use ${passMaxLength.value} or fewer characters`;
    }

    const score = scorePassword();
    if (score > 90) {
        return PASSWORD_STRENGTH.veryStrong;
    }
    if (score > 70) {
        return PASSWORD_STRENGTH.strong;
    }
    if (score > 45) {
        return PASSWORD_STRENGTH.good;
    }

    return PASSWORD_STRENGTH.weak;
});

/**
 * Color for indicator between red as weak and green as strong password.
 */
const passwordStrengthColor = computed((): string => {
    return PASSWORD_STRENGTH_COLORS[passwordStrength.value] || PASSWORD_STRENGTH_COLORS.default;
});

/**
 * Fills password strength indicator bar.
 */
const barWidth = computed((): string => {
    return BAR_WIDTH[passwordStrength.value] || BAR_WIDTH.default;
});

const strengthLabelColor = computed((): { color: string } => {
    return { color: passwordStrengthColor.value };
});

const hasLowerAndUpperCaseLetters = computed((): boolean => {
    return /[a-z]/.test(props.password) && /[A-Z]/.test(props.password);
});

const hasSpecialCharacter = computed((): boolean => {
    return /\W/.test(props.password);
});

/**
 * Returns password strength score depends on length, case variations and special characters.
 */
function scorePassword(): number {
    const password = props.password;
    let score = 0;

    const letters: number[] = [];
    for (let i = 0; i < password.length; i++) {
        letters[password[i]] = (letters[password[i]] || 0) + 1;
        score += 5 / letters[password[i]];
    }
    score = Math.min(score, 60);

    const variations: boolean[] = [
        /\d/.test(password),
        /[a-z]/.test(password),
        /[A-Z]/.test(password),
        /\W/.test(password),
    ];

    score += variations.filter(Boolean).length * 10;

    const isSequential = (str: string): boolean => {
        const sequences = 'abcdefghijklmnopqrstuvwxyz0123456789';
        const reversed = sequences.split('').reverse().join('');

        for (let i = 0; i < str.length - 2; i++) {
            const substr = str.slice(i, i + 3);
            if (sequences.includes(substr) || reversed.includes(substr)) {
                return true;
            }
        }
        return false;
    };

    if (isSequential(password.toLowerCase())) {
        // Penalize sequential patterns
        score -= 20;
    }

    if (props.email && password === props.email) {
        // Penalize password that is the same as email
        score -= 20;
    }

    return Math.max(score, 0);
}
</script>

<style scoped lang="scss">
.positioning {
    z-index: 1;
}
</style>
