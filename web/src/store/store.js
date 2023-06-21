import { configureStore } from '@reduxjs/toolkit'
import rootReducer from './reducer';
import storage from 'redux-persist/lib/storage';
import { persistReducer,persistStore } from 'redux-persist';
import thunk from 'redux-thunk';
import { logger } from 'redux-logger';


// for Redux Persist
const persistConfig = {
  key: 'pacui-root',
  storage: storage
};

const persistedReducer = persistReducer(persistConfig, rootReducer);

export const store = configureStore({
  reducer : persistedReducer,
  middleware:[thunk,logger]
});

export const persistor = persistStore(store);

