'use client';

import { capitalCase } from 'change-case';
import pluralize from 'pluralize';
import { useCallback } from 'react';

export function useUserFriendlyName(): (
  name: string | undefined,
  options?: 'plural' | 'singular' | { plural?: boolean; singular?: boolean }
) => string {
  return useCallback(
    (
      name: string | undefined,
      options?: 'plural' | 'singular' | { plural?: boolean; singular?: boolean }
    ): string => {
      if (!name) {
        return '';
      }

      const humanized = capitalCase(name);

      if (options === 'plural') {
        return pluralize(humanized);
      }

      if (options === 'singular') {
        return pluralize.singular(humanized);
      }

      if (typeof options === 'object') {
        if (options?.plural) {
          return pluralize(humanized);
        }
        if (options?.singular) {
          return pluralize.singular(humanized);
        }
      }

      return humanized;
    },
    []
  );
}
